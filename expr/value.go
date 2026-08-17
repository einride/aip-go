package expr

import (
	"math"
	"strconv"
	"time"
)

// ValueKind identifies the type stored in a Value.
type ValueKind uint8

const (
	NullKind ValueKind = iota
	BoolKind
	Int64Kind
	Uint64Kind
	Float64Kind
	StringKind
	BytesKind
	TimeKind
	DurationKind
)

// Value is a typed filter literal or cursor value.
//
// Values are immutable and comparable. Byte slices are copied on input and
// output.
//
// The payload is packed rather than kept in a field per kind: text holds the
// string or byte payload, num holds every fixed-width kind, and nsec extends
// num for times. A time is stored as seconds since the Unix epoch plus
// nanoseconds, which spans the whole time.Time range; nanoseconds alone would
// only reach 1678 through 2262.
type Value struct {
	text string    // StringKind text, BytesKind payload
	num  uint64    // bool, int64, uint64, float64 bits, duration, time seconds, BytesKind non-nil flag
	nsec uint32    // TimeKind nanoseconds
	kind ValueKind //
}

func NullValue() Value { return Value{kind: NullKind} }

func BoolValue(v bool) Value {
	var num uint64
	if v {
		num = 1
	}
	return Value{kind: BoolKind, num: num}
}

func Int64Value(v int64) Value            { return Value{kind: Int64Kind, num: uint64(v)} }
func Uint64Value(v uint64) Value          { return Value{kind: Uint64Kind, num: v} }
func Float64Value(v float64) Value        { return Value{kind: Float64Kind, num: math.Float64bits(v)} }
func StringValue(v string) Value          { return Value{kind: StringKind, text: v} }
func DurationValue(v time.Duration) Value { return Value{kind: DurationKind, num: uint64(v)} }

func TimeValue(v time.Time) Value {
	utc := v.UTC()
	return Value{kind: TimeKind, num: uint64(utc.Unix()), nsec: uint32(utc.Nanosecond())}
}

// BytesValue copies v. A nil slice stays nil on the way out, which keeps the
// distinction a SQL driver makes between a null and an empty blob.
func BytesValue(v []byte) Value {
	if v == nil {
		return Value{kind: BytesKind}
	}
	return Value{kind: BytesKind, num: 1, text: string(v)}
}

func (v Value) Kind() ValueKind { return v.kind }

// Any returns the underlying Go value. Bytes are returned as a copy.
func (v Value) Any() any {
	switch v.kind {
	case BoolKind:
		return v.num != 0
	case Int64Kind:
		return int64(v.num)
	case Uint64Kind:
		return v.num
	case Float64Kind:
		return math.Float64frombits(v.num)
	case StringKind:
		return v.text
	case BytesKind:
		if v.num == 0 {
			return []byte(nil)
		}
		return []byte(v.text)
	case TimeKind:
		return v.utc()
	case DurationKind:
		return time.Duration(v.num)
	default: // NullKind and invalid internal values.
		return nil
	}
}

// Valid reports whether v holds a representable value. NaN, infinities, and
// the zero time are rejected.
func (v Value) Valid() bool {
	if v.kind == Float64Kind {
		f := math.Float64frombits(v.num)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return false
		}
	}
	if v.kind == TimeKind && v.utc().IsZero() {
		return false
	}
	return v.kind <= DurationKind
}

// Canonical returns the AIP-160 literal representation of v.
//
// It is deterministic, which is what the cursor fingerprint relies on, but it
// does not preserve type: an integral float renders without a decimal point,
// so re-parsing "1" yields an integer rather than the float it came from.
// Canonical output is hashed, never fed back through the parser.
func (v Value) Canonical() string {
	switch v.kind {
	case BoolKind:
		return strconv.FormatBool(v.num != 0)
	case Int64Kind:
		return strconv.FormatInt(int64(v.num), 10)
	case Uint64Kind:
		return strconv.FormatUint(v.num, 10)
	case Float64Kind:
		return strconv.FormatFloat(math.Float64frombits(v.num), 'g', -1, 64)
	case StringKind:
		return strconv.Quote(v.text)
	case BytesKind:
		return strconv.Quote(v.text)
	case TimeKind:
		return v.utc().Format(time.RFC3339Nano)
	case DurationKind:
		return time.Duration(v.num).String()
	default: // NullKind and invalid internal values.
		return "null"
	}
}

// utc rebuilds the time stored in v. Only meaningful for TimeKind.
func (v Value) utc() time.Time {
	return time.Unix(int64(v.num), int64(v.nsec)).UTC()
}
