package unbuilt

import "testing"

// The IMPLICIT constraint: no directive here at all, only the GOOS in the
// filename, which go/build honours identically. Neither this fleet's darwin
// machines nor its linux CI ever compiles it.
func FuzzPlan9(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
