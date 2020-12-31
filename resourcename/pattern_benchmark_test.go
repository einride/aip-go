package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

// nolint: gochecknoglobals
var (
	nameSink string
	errSink  error
)

func BenchmarkPattern_MarshalResourceName(b *testing.B) {
	pattern, err := ParsePattern("publishers/{publisher}/books/{book}")
	assert.NilError(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		name, _ := pattern.MarshalResourceName("1", "2")
		nameSink = name
	}
}

func BenchmarkPattern_ValidateResourceName(b *testing.B) {
	pattern, err := ParsePattern("publishers/{publisher}/books/{book}")
	assert.NilError(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		errSink = pattern.ValidateResourceName("publishers/1/books/2")
	}
}
