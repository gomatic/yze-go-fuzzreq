package unbuilt

import "testing"

// An ordinary .go file. `go list` names it in GoFiles and `go test -fuzz`
// cannot reach a target declared outside a test file at all.
func FuzzInSource(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
