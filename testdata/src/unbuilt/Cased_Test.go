package unbuilt

import "testing"

// The go tool's test-file rule is CASE-SENSITIVE, so this file is ordinary
// compiled source and the declaration below is never a fuzz target — it is one
// shift key from being one, on a filesystem that opens it under either
// spelling.
func FuzzCased(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
