//go:build never_ever_built

package unbuilt

import "testing"

// A constraint no ordinary build satisfies. `go list` puts this file in
// IgnoredGoFiles, `go test ./...` never compiles it, and CI never runs it.
func FuzzTagged(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
