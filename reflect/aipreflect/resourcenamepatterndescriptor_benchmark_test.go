package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

// nolint: gochecknoglobals
var (
	nameSink string
	errSink  error
)

func BenchmarkResourceNamePatternDescriptor_MarshalResourceName(b *testing.B) {
	desc, err := NewResourceNamePatternDescriptor("publishers/{publisher}/books/{book}")
	assert.NilError(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		name, _ := desc.MarshalResourceName("1", "2")
		nameSink = name
	}
}

func BenchmarkResourceNamePatternDescriptor_ValidateResourceName(b *testing.B) {
	desc, err := NewResourceNamePatternDescriptor("publishers/{publisher}/books/{book}")
	assert.NilError(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		errSink = desc.ValidateResourceName("publishers/1/books/2")
	}
}
