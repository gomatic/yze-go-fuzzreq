package unbuilt

import "testing"

// A leading DOT: the go tool ignores the file entirely, so this target is never
// compiled and never run. `go list` does not name it in TestGoFiles.
func FuzzDotted(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
