package unbuilt

import "testing"

// The name CONTAINS `_test.go` and does not end in it, so `go list` puts this
// file in GoFiles and the go tool never runs the target below. The judged side
// of the rule has a fixture for this widening; until this one existed the
// EVIDENCE side did not, and relaxing its suffix to a substring exempted the
// package while `go test` printed "[no test files]".
func FuzzGolden(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
