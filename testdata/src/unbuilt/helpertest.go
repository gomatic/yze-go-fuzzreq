package unbuilt

import "testing"

// The LEFT EDGE of the literal: this name ends in `test.go` and not in
// `_test.go`, so `go list` puts it in GoFiles and `go test -fuzz` cannot reach
// the declaration below. net/http/httptest/httptest.go wears the same name.
func FuzzHelper(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
