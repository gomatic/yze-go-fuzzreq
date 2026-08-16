package unbuilt

import "testing"

// A leading UNDERSCORE: ignored by the go tool for the same reason.
func FuzzUnderscored(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = Parse(raw) })
}
