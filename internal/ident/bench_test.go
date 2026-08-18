package ident_test

import (
	"testing"

	"github.com/ysomad/flop/internal/ident"
)

// BenchmarkValidPath guards against a return to rune-slice scanning. The rule
// is ASCII-only, so it must not allocate.
func BenchmarkValidPath(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if !ident.ValidPath("orders.created_at") {
			b.Fatal("want a valid path")
		}
	}
}
