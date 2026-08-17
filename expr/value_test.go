package expr

import (
	"math"
	"testing"
	"time"
	"unsafe"

	"github.com/ysomad/flop/internal/assert"
)

func TestValueSize(t *testing.T) {
	t.Parallel()

	if got := unsafe.Sizeof(Value{}); got > 32 {
		t.Fatalf("Value grew to %d bytes, want at most 32", got)
	}
}

func TestNullValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{}, NullValue())
}

func TestBoolValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{kind: BoolKind}, BoolValue(false))
	assert.Equal(t, Value{kind: BoolKind, num: 1}, BoolValue(true))
}

func TestInt64Value(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{kind: Int64Kind, num: 1 << 63}, Int64Value(math.MinInt64))
	assert.Equal(t, Value{kind: Int64Kind}, Int64Value(0))
	assert.Equal(t, Value{kind: Int64Kind, num: math.MaxInt64}, Int64Value(math.MaxInt64))
}

func TestUint64Value(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{kind: Uint64Kind}, Uint64Value(0))
	assert.Equal(t, Value{kind: Uint64Kind, num: math.MaxUint64}, Uint64Value(math.MaxUint64))
}

func TestFloat64Value(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{kind: Float64Kind, num: math.Float64bits(-12.5)}, Float64Value(-12.5))
	assert.Equal(t, Value{kind: Float64Kind}, Float64Value(0))
	assert.Equal(t, Value{kind: Float64Kind, num: math.Float64bits(math.MaxFloat64)}, Float64Value(math.MaxFloat64))
}

func TestStringValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Value{kind: StringKind}, StringValue(""))
	assert.Equal(t, Value{kind: StringKind, text: "hello, 世界"}, StringValue("hello, 世界"))
}

func TestTimeValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "offset zone",
			input: time.Date(2026, time.August, 16, 12, 30, 45, 123, time.FixedZone("UTC+7", 7*60*60)),
		},
		{name: "before the nanosecond epoch range", input: time.Date(1000, time.March, 4, 5, 6, 7, 8, time.UTC)},
		{name: "after the nanosecond epoch range", input: time.Date(3000, time.March, 4, 5, 6, 7, 8, time.UTC)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			utc := test.input.UTC()
			got := TimeValue(test.input)

			assert.Equal(t, Value{kind: TimeKind, num: uint64(utc.Unix()), nsec: uint32(utc.Nanosecond())}, got)
			assert.Equal(t, any(utc), got.Any())
			assert.Equal(t, utc.Format(time.RFC3339Nano), got.Canonical())
			assert.Equal(t, true, got.Valid())
		})
	}
}

func TestDurationValue(t *testing.T) {
	t.Parallel()

	negative := -3 * time.Second
	assert.Equal(t, Value{kind: DurationKind, num: uint64(negative)}, DurationValue(negative))
	assert.Equal(t, Value{kind: DurationKind}, DurationValue(0))
	assert.Equal(t, Value{kind: DurationKind, num: uint64(3 * time.Second)}, DurationValue(3*time.Second))
}

func TestBytesValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		want    Value
		wantAny any
	}{
		// A nil slice stays nil, which a SQL driver binds as null rather than
		// as an empty blob.
		{name: "nil", input: nil, want: Value{kind: BytesKind}, wantAny: []byte(nil)},
		{name: "empty", input: []byte{}, want: Value{kind: BytesKind, num: 1}, wantAny: []byte{}},
		{
			name:    "value",
			input:   []byte("input"),
			want:    Value{kind: BytesKind, num: 1, text: "input"},
			wantAny: []byte("input"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := BytesValue(test.input)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantAny, got.Any())
		})
	}

	input := []byte("input")
	got := BytesValue(input)
	input[0] = 'X'

	assert.Equal(t, Value{kind: BytesKind, num: 1, text: "input"}, got)
}

func TestValue_Kind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  Value
		want ValueKind
	}{
		{name: "null", val: NullValue(), want: NullKind},
		{name: "bool", val: BoolValue(true), want: BoolKind},
		{name: "int64", val: Int64Value(1), want: Int64Kind},
		{name: "uint64", val: Uint64Value(1), want: Uint64Kind},
		{name: "float64", val: Float64Value(1), want: Float64Kind},
		{name: "string", val: StringValue("value"), want: StringKind},
		{name: "bytes", val: BytesValue([]byte("value")), want: BytesKind},
		{name: "time", val: TimeValue(time.Unix(1, 0)), want: TimeKind},
		{name: "duration", val: DurationValue(time.Second), want: DurationKind},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.val.Kind())
		})
	}
}

func TestValue_Any(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value Value
		want  any
	}{
		{name: "null", value: NullValue(), want: nil},
		{name: "bool", value: BoolValue(true), want: true},
		{name: "int64", value: Int64Value(-1), want: int64(-1)},
		{name: "uint64", value: Uint64Value(1), want: uint64(1)},
		{name: "float64", value: Float64Value(1.5), want: 1.5},
		{name: "string", value: StringValue("value"), want: "value"},
		{name: "bytes", value: BytesValue([]byte("value")), want: []byte("value")},
		{name: "time", value: TimeValue(time.Unix(1, 0)), want: time.Unix(1, 0).UTC()},
		{name: "duration", value: DurationValue(time.Second), want: time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.value.Any())
		})
	}

	bytesValue := BytesValue([]byte("value"))
	got := bytesValue.Any().([]byte)
	got[0] = 'X'
	assert.Equal(t, "value", bytesValue.text)
}
