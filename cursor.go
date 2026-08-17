package flop

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"slices"
	"time"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/ident"
)

const (
	cursorVersion  = 1
	maxCursorBytes = 64 << 10
)

// CursorField is one field value in a cursor position.
type CursorField struct {
	Path  string
	Value expr.Value
}

// CursorPosition is an ordered tuple that identifies a row under the effective
// storage ordering.
type CursorPosition []CursorField

// CursorParams contains the cursor parameters a storage implementation needs.
//
// Limit reports one surplus row so NewCursorPage can determine whether a next
// page exists. After reports the position the page starts after.
type CursorParams struct {
	after    CursorPosition
	pageSize int32
	binding  [sha256.Size]byte
	orderBy  OrderBy
	filter   Filter
}

// After returns the position the next page starts after. It is empty for the
// first page.
func (p CursorParams) After() CursorPosition { return p.after.clone() }

// Limit is the page size plus one surplus row, saturated at math.MaxInt32.
func (p CursorParams) Limit() int32 {
	if p.pageSize == math.MaxInt32 {
		return math.MaxInt32
	}
	return p.pageSize + 1
}

// NewCursorParams parses pagination inputs and returns backend-neutral storage
// parameters. orderBy and filter bind a cursor to the current request.
// pageSize must already be normalized by the caller.
func NewCursorParams(
	cursor string,
	pageSize int32,
	orderBy OrderBy,
	f Filter,
) (CursorParams, error) {
	if pageSize <= 0 {
		return CursorParams{}, fmt.Errorf("flop: page size must be positive")
	}
	binding := cursorBinding(orderBy, f)
	params := CursorParams{
		pageSize: pageSize,
		binding:  binding,
		orderBy:  orderBy.clone(),
		filter:   f,
	}
	if cursor == "" {
		return params, nil
	}
	after, gotBinding, err := decodeCursor(cursor)
	if err != nil {
		return CursorParams{}, err
	}
	if binding != gotBinding {
		return CursorParams{}, fmt.Errorf("flop: cursor does not match the request")
	}
	params.after = after.clone()
	return params, nil
}

// OrderBy returns the resolved ordering associated with these parameters.
func (p CursorParams) OrderBy() OrderBy { return p.orderBy.clone() }

// Filter returns the validated filter associated with these parameters.
func (p CursorParams) Filter() Filter { return p.filter }

// CursorPage is a page returned by cursor pagination.
type CursorPage[T any] struct {
	Items      []T
	NextCursor string
}

// NewCursorPage trims the surplus row and creates a cursor from the last
// returned row when another page exists.
func NewCursorPage[T any](
	rows []T,
	params CursorParams,
	position func(T) (CursorPosition, error),
) (CursorPage[T], error) {
	if params.pageSize <= 0 {
		return CursorPage[T]{}, fmt.Errorf("flop: invalid cursor parameters")
	}
	if int64(len(rows)) > int64(params.Limit()) {
		return CursorPage[T]{}, fmt.Errorf("flop: storage returned too many rows")
	}
	hasNext := int64(len(rows)) > int64(params.pageSize)
	items := rows
	if hasNext {
		items = rows[:params.pageSize]
	}
	page := CursorPage[T]{Items: items}
	if !hasNext {
		return page, nil
	}
	if position == nil {
		return CursorPage[T]{}, fmt.Errorf("flop: nil cursor position function")
	}
	next, err := position(items[len(items)-1])
	if err != nil {
		return CursorPage[T]{}, fmt.Errorf("flop: create cursor position: %w", err)
	}
	if err := next.validate(); err != nil {
		return CursorPage[T]{}, err
	}
	page.NextCursor = encodeCursor(next, params.binding)
	return page, nil
}

func (p CursorPosition) validate() error {
	if len(p) == 0 {
		return fmt.Errorf("flop: cursor position must not be empty")
	}
	seen := make(map[string]struct{}, len(p))
	for _, field := range p {
		if !ident.ValidPath(field.Path) {
			return fmt.Errorf("flop: invalid cursor position path")
		}
		if _, ok := seen[field.Path]; ok {
			return fmt.Errorf("flop: duplicate cursor position path")
		}
		if !field.Value.Valid() {
			return fmt.Errorf("flop: invalid cursor position value")
		}
		// A null is representable, so Valid accepts it, but no backend can
		// build a seek predicate from one.
		if field.Value.Kind() == expr.NullKind {
			return fmt.Errorf("flop: cursor position value must not be null")
		}
		seen[field.Path] = struct{}{}
	}
	return nil
}

// clone copies the position. Values are immutable, so a shallow copy is enough.
func (p CursorPosition) clone() CursorPosition {
	return slices.Clone(p)
}

func cursorBinding(o OrderBy, f Filter) [sha256.Size]byte {
	hash := sha256.New()
	writeHashString(hash, o.String())
	writeHashString(hash, f.String())
	var bb [sha256.Size]byte
	copy(bb[:], hash.Sum(nil))
	return bb
}

func writeHashString(w io.Writer, s string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(s)))
	_, _ = w.Write(size[:])
	_, _ = io.WriteString(w, s)
}

func encodeCursor(pos CursorPosition, bb [sha256.Size]byte) string {
	var buf bytes.Buffer
	buf.WriteByte(cursorVersion)
	buf.Write(bb[:])
	writeUvarint(&buf, uint64(len(pos)))
	for _, field := range pos {
		writeBytes(&buf, []byte(field.Path))
		writeValue(&buf, field.Value)
	}

	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func decodeCursor(cursor string) (CursorPosition, [sha256.Size]byte, error) {
	var zero, binding [sha256.Size]byte
	if len(cursor) > base64.RawURLEncoding.EncodedLen(maxCursorBytes) {
		return nil, zero, fmt.Errorf("flop: cursor is too large")
	}
	encoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, zero, fmt.Errorf("flop: invalid cursor")
	}
	reader := bytes.NewReader(encoded)
	version, err := reader.ReadByte()
	if err != nil || version != cursorVersion {
		return nil, zero, fmt.Errorf("flop: unsupported cursor")
	}
	if _, err := io.ReadFull(reader, binding[:]); err != nil {
		return nil, zero, fmt.Errorf("flop: invalid cursor")
	}
	count, err := binary.ReadUvarint(reader)
	if err != nil || count > uint64(reader.Len()) {
		return nil, zero, fmt.Errorf("flop: invalid cursor")
	}
	position := make(CursorPosition, int(count))
	for i := range position {
		path, err := readBytes(reader)
		if err != nil {
			return nil, zero, err
		}
		val, err := readValue(reader)
		if err != nil {
			return nil, zero, err
		}
		position[i] = CursorField{Path: string(path), Value: val}
	}
	if reader.Len() != 0 {
		return nil, zero, fmt.Errorf("flop: invalid cursor")
	}
	if err := position.validate(); err != nil {
		return nil, zero, err
	}
	return position, binding, nil
}

func writeUvarint(w io.Writer, v uint64) {
	var buf [binary.MaxVarintLen64]byte
	size := binary.PutUvarint(buf[:], v)
	_, _ = w.Write(buf[:size])
}

func writeBytes(w io.Writer, bb []byte) {
	writeUvarint(w, uint64(len(bb)))
	_, _ = w.Write(bb)
}

func readBytes(r *bytes.Reader) ([]byte, error) {
	size, err := binary.ReadUvarint(r)
	if err != nil || size > uint64(r.Len()) {
		return nil, fmt.Errorf("flop: invalid cursor")
	}
	v := make([]byte, int(size))
	_, _ = io.ReadFull(r, v)
	return v, nil
}

// writeValue encodes v as a kind byte followed by its payload. The payload
// layout is selected by the dynamic type of Any, which is exhaustive over every
// non-null kind; a null value writes no payload.
func writeValue(w io.Writer, v expr.Value) {
	_, _ = w.Write([]byte{byte(v.Kind())})
	var fixed [8]byte
	switch value := v.Any().(type) {
	case bool:
		if value {
			_, _ = w.Write([]byte{1})
		} else {
			_, _ = w.Write([]byte{0})
		}
	case int64:
		binary.BigEndian.PutUint64(fixed[:], uint64(value))
		_, _ = w.Write(fixed[:])
	case uint64:
		binary.BigEndian.PutUint64(fixed[:], value)
		_, _ = w.Write(fixed[:])
	case float64:
		binary.BigEndian.PutUint64(fixed[:], math.Float64bits(value))
		_, _ = w.Write(fixed[:])
	case string:
		writeBytes(w, []byte(value))
	case []byte:
		writeBytes(w, value)
	case time.Time:
		writeBytes(w, []byte(value.UTC().Format(time.RFC3339Nano)))
	case time.Duration:
		binary.BigEndian.PutUint64(fixed[:], uint64(value))
		_, _ = w.Write(fixed[:])
	}
}

func readValue(r *bytes.Reader) (expr.Value, error) {
	kind, err := r.ReadByte()
	if err != nil {
		return expr.Value{}, fmt.Errorf("flop: invalid cursor")
	}
	var fixed [8]byte
	readFixed := func() error {
		if _, err := io.ReadFull(r, fixed[:]); err != nil {
			return fmt.Errorf("flop: invalid cursor")
		}
		return nil
	}
	switch expr.ValueKind(kind) {
	case expr.NullKind:
		return expr.NullValue(), nil
	case expr.BoolKind:
		value, err := r.ReadByte()
		if err != nil || value > 1 {
			return expr.Value{}, fmt.Errorf("flop: invalid cursor")
		}
		return expr.BoolValue(value == 1), nil
	case expr.Int64Kind:
		if err := readFixed(); err != nil {
			return expr.Value{}, err
		}
		return expr.Int64Value(int64(binary.BigEndian.Uint64(fixed[:]))), nil
	case expr.Uint64Kind:
		if err := readFixed(); err != nil {
			return expr.Value{}, err
		}
		return expr.Uint64Value(binary.BigEndian.Uint64(fixed[:])), nil
	case expr.Float64Kind:
		if err := readFixed(); err != nil {
			return expr.Value{}, err
		}
		value := expr.Float64Value(math.Float64frombits(binary.BigEndian.Uint64(fixed[:])))
		if !value.Valid() {
			return expr.Value{}, fmt.Errorf("flop: invalid cursor")
		}
		return value, nil
	case expr.StringKind:
		value, err := readBytes(r)
		return expr.StringValue(string(value)), err
	case expr.BytesKind:
		value, err := readBytes(r)
		return expr.BytesValue(value), err
	case expr.TimeKind:
		value, err := readBytes(r)
		if err != nil {
			return expr.Value{}, err
		}
		parsed, err := time.Parse(time.RFC3339Nano, string(value))
		if err != nil {
			return expr.Value{}, fmt.Errorf("flop: invalid cursor")
		}
		return expr.TimeValue(parsed), nil
	case expr.DurationKind:
		if err := readFixed(); err != nil {
			return expr.Value{}, err
		}
		return expr.DurationValue(time.Duration(binary.BigEndian.Uint64(fixed[:]))), nil
	default:
		return expr.Value{}, fmt.Errorf("flop: invalid cursor")
	}
}
