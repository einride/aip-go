package flop_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

func TestNewCursorParams(t *testing.T) {
	t.Parallel()
	orderBy, err := flop.ParseOrderBy("created_at desc, id")
	assert.NoError(t, err)
	activeFilter, err := flop.ParseFilter("active = true")
	assert.NoError(t, err)
	first, err := flop.NewCursorParams("", 1, orderBy, activeFilter)
	assert.NoError(t, err)
	createdAt := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)
	position := flop.CursorPosition{
		{Path: "created_at", Value: expr.TimeValue(createdAt)},
		{Path: "id", Value: expr.StringValue("users/1")},
	}
	page, err := flop.NewCursorPage(
		[]int32{1, 2},
		first,
		func(int32) (flop.CursorPosition, error) { return position, nil },
	)
	assert.NoError(t, err)
	// A null is not a valid position value, so every other kind is covered here
	// and the null rejection is exercised by TestNewCursorPage.
	allValues := flop.CursorPosition{
		{Path: "bool_false", Value: expr.BoolValue(false)},
		{Path: "bool_true", Value: expr.BoolValue(true)},
		{Path: "int_value", Value: expr.Int64Value(math.MinInt64)},
		{Path: "uint_value", Value: expr.Uint64Value(math.MaxUint64)},
		{Path: "float_value", Value: expr.Float64Value(1.5)},
		{Path: "string_value", Value: expr.StringValue("value")},
		{Path: "bytes_value", Value: expr.BytesValue([]byte("value"))},
		{Path: "time_value", Value: expr.TimeValue(createdAt)},
		{Path: "duration_value", Value: expr.DurationValue(-time.Second)},
	}
	allValuesPage, err := flop.NewCursorPage(
		[]int32{1, 2},
		first,
		func(int32) (flop.CursorPosition, error) { return allValues, nil },
	)
	assert.NoError(t, err)
	otherFilter, err := flop.ParseFilter("active = false")
	assert.NoError(t, err)
	otherOrderBy, err := flop.ParseOrderBy("id")
	assert.NoError(t, err)
	encode := func(plaintext []byte) string { return base64.RawURLEncoding.EncodeToString(plaintext) }
	shortPath := cursorPlaintext(1)
	shortPath = append(shortPath, 5, 'a')
	trailing := append(cursorPlaintext(0), 0)
	invalidCount := append(cursorPrefix(), 0x80)
	invalidFloat := make([]byte, 8)
	binary.BigEndian.PutUint64(invalidFloat, math.Float64bits(math.NaN()))

	type args struct {
		cursor  string
		size    int32
		orderBy flop.OrderBy
		filter  flop.Filter
	}
	type cursorParams struct {
		limit   int32
		after   flop.CursorPosition
		orderBy flop.OrderBy
		filter  flop.Filter
	}
	tests := []struct {
		name             string
		args             args
		wantCursorParams cursorParams
		wantErr          assert.ErrorFunc
	}{
		{
			name:             "page size",
			args:             args{size: 2},
			wantCursorParams: cursorParams{limit: 3},
			wantErr:          assert.NoError,
		},
		{
			name:             "maximum int32 page size",
			args:             args{size: math.MaxInt32},
			wantCursorParams: cursorParams{limit: math.MaxInt32},
			wantErr:          assert.NoError,
		},
		{
			name: "cursor",
			args: args{
				cursor: page.NextCursor, size: 2,
				orderBy: orderBy, filter: activeFilter,
			},
			wantCursorParams: cursorParams{
				limit: 3, after: position, orderBy: orderBy, filter: activeFilter,
			},
			wantErr: assert.NoError,
		},
		{
			name: "all cursor value kinds",
			args: args{
				cursor: allValuesPage.NextCursor, size: 1,
				orderBy: orderBy, filter: activeFilter,
			},
			wantCursorParams: cursorParams{
				limit: 2, after: allValues, orderBy: orderBy, filter: activeFilter,
			},
			wantErr: assert.NoError,
		},
		{
			name: "page size can change between requests",
			args: args{
				cursor: page.NextCursor, size: 3,
				orderBy: orderBy, filter: activeFilter,
			},
			wantCursorParams: cursorParams{
				limit: 4, after: position, orderBy: orderBy, filter: activeFilter,
			},
			wantErr: assert.NoError,
		},
		{
			name: "changed order",
			args: args{
				cursor: page.NextCursor, size: 2,
				orderBy: otherOrderBy, filter: activeFilter,
			},
			wantErr: assert.Error,
		},
		{
			name: "changed filter",
			args: args{
				cursor: page.NextCursor, size: 2,
				orderBy: orderBy, filter: otherFilter,
			},
			wantErr: assert.Error,
		},
		{
			name:    "cursor too large",
			args:    args{cursor: strings.Repeat("A", base64.RawURLEncoding.EncodedLen(64<<10)+1), size: 1},
			wantErr: assert.Error,
		},
		{name: "invalid base64 cursor", args: args{cursor: "*", size: 1}, wantErr: assert.Error},
		{
			name:    "empty cursor payload",
			args:    args{cursor: base64.RawURLEncoding.EncodeToString([]byte{0}), size: 1},
			wantErr: assert.Error,
		},
		{name: "unsupported cursor version", args: args{cursor: encode([]byte{2}), size: 1}, wantErr: assert.Error},
		{name: "short cursor binding", args: args{cursor: encode([]byte{1}), size: 1}, wantErr: assert.Error},
		{
			name:    "invalid cursor position count",
			args:    args{cursor: encode(invalidCount), size: 1},
			wantErr: assert.Error,
		},
		{
			name:    "cursor position count exceeds payload",
			args:    args{cursor: encode(cursorPlaintext(1)), size: 1},
			wantErr: assert.Error,
		},
		{
			name:    "invalid cursor position path length",
			args:    args{cursor: encode(shortPath), size: 1},
			wantErr: assert.Error,
		},
		{
			name:    "missing cursor position value",
			args:    args{cursor: encode(cursorPlaintext(1, cursorField("id", nil))), size: 1},
			wantErr: assert.Error,
		},
		{name: "trailing cursor data", args: args{cursor: encode(trailing), size: 1}, wantErr: assert.Error},
		{name: "empty cursor position", args: args{cursor: encode(cursorPlaintext(0)), size: 1}, wantErr: assert.Error},
		{
			name: "invalid path",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("invalid-path", []byte{byte(expr.NullKind)}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "duplicate path",
			args: args{cursor: encode(cursorPlaintext(2,
				cursorField("id", []byte{byte(expr.NullKind)}),
				cursorField("id", []byte{byte(expr.NullKind)}),
			)), size: 1},
			wantErr: assert.Error,
		},
		{
			name: "invalid bool",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.BoolKind), 2}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "short int64",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.Int64Kind), 0}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "short uint64",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.Uint64Kind), 0}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "short float64",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.Float64Kind), 0}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "non-finite float64",
			args: args{
				cursor: encode(
					cursorPlaintext(
						1,
						cursorField("v", append([]byte{byte(expr.Float64Kind)}, invalidFloat...)),
					),
				),
				size: 1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid string length",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.StringKind), 2, 'a'}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid bytes length",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.BytesKind), 2, 'a'}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid time length",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.TimeKind), 2, 'a'}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid time",
			args: args{
				cursor: encode(
					cursorPlaintext(1, cursorField("v", []byte{byte(expr.TimeKind), 3, 'b', 'a', 'd'})),
				),
				size: 1,
			},
			wantErr: assert.Error,
		},
		{
			name: "short duration",
			args: args{
				cursor: encode(cursorPlaintext(1, cursorField("v", []byte{byte(expr.DurationKind), 0}))),
				size:   1,
			},
			wantErr: assert.Error,
		},
		{
			name:    "unknown value kind",
			args:    args{cursor: encode(cursorPlaintext(1, cursorField("v", []byte{math.MaxUint8}))), size: 1},
			wantErr: assert.Error,
		},
		{name: "zero page size", wantErr: assert.Error},
		{
			name:    "negative page size",
			args:    args{size: -1},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := flop.NewCursorParams(
				test.args.cursor,
				test.args.size,
				test.args.orderBy,
				test.args.filter,
			)
			test.wantErr(t, gotErr)
			if gotErr != nil {
				return
			}
			assert.Equal(t, test.wantCursorParams, cursorParams{
				limit:   got.Limit(),
				after:   got.After(),
				orderBy: got.OrderBy(),
				filter:  got.Filter(),
			})
		})
	}
}

func TestNewCursorPage(t *testing.T) {
	params, err := flop.NewCursorParams("", 2, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	positionErr := errors.New("cursor position failed")

	type args struct {
		rows     []int32
		params   flop.CursorParams
		position func(int32) (flop.CursorPosition, error)
	}
	type cursorPage struct {
		items     []int32
		hasCursor bool
	}
	tests := []struct {
		name           string
		args           args
		wantCursorPage cursorPage
		wantErr        assert.ErrorFunc
	}{
		{
			name:    "zero parameters",
			args:    args{rows: []int32{}},
			wantErr: assert.Error,
		},
		{
			name:           "last page",
			args:           args{rows: []int32{1, 2}, params: params},
			wantCursorPage: cursorPage{items: []int32{1, 2}},
			wantErr:        assert.NoError,
		},
		{
			name:           "empty page",
			args:           args{rows: []int32{}, params: params},
			wantCursorPage: cursorPage{items: []int32{}},
			wantErr:        assert.NoError,
		},
		{
			name: "next page",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(row int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "id", Value: expr.Int64Value(int64(row))}}, nil
				},
			},
			wantCursorPage: cursorPage{items: []int32{1, 2}, hasCursor: true},
			wantErr:        assert.NoError,
		},
		{
			name: "nil cursor position function",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
			},
			wantErr: assert.Error,
		},
		{
			name: "cursor position function error",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return nil, positionErr
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "too many rows",
			args: args{
				rows: []int32{1, 2, 3, 4}, params: params,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid cursor position",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "null cursor position value",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "id", Value: expr.NullValue()}}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "duplicate cursor position path",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{
						{Path: "id", Value: expr.Int64Value(1)},
						{Path: "id", Value: expr.Int64Value(2)},
					}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "non-finite cursor position value",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "score", Value: expr.Float64Value(math.NaN())}}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "zero cursor position time",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "created_at", Value: expr.TimeValue(time.Time{})}}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid cursor position path",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "invalid-path", Value: expr.Int64Value(1)}}, nil
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "empty cursor position path",
			args: args{
				rows: []int32{1, 2, 3}, params: params,
				position: func(int32) (flop.CursorPosition, error) {
					return flop.CursorPosition{{Path: "", Value: expr.Int64Value(1)}}, nil
				},
			},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := flop.NewCursorPage(test.args.rows, test.args.params, test.args.position)
			assert.Equal(t, test.wantCursorPage, cursorPage{
				items:     got.Items,
				hasCursor: got.NextCursor != "",
			})
			test.wantErr(t, gotErr)
		})
	}
}

func cursorPrefix() []byte {
	prefix := []byte{1}
	return append(prefix, make([]byte, sha256.Size)...)
}

func cursorPlaintext(count uint64, fields ...[]byte) []byte {
	plaintext := cursorPrefix()
	plaintext = appendUvarint(plaintext, count)
	for _, field := range fields {
		plaintext = append(plaintext, field...)
	}
	return plaintext
}

func cursorField(path string, value []byte) []byte {
	field := appendUvarint(nil, uint64(len(path)))
	field = append(field, path...)
	return append(field, value...)
}

func appendUvarint(dst []byte, value uint64) []byte {
	var encoded [binary.MaxVarintLen64]byte
	size := binary.PutUvarint(encoded[:], value)
	return append(dst, encoded[:size]...)
}
