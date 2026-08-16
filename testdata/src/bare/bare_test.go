package bare

import "testing"

// Fuzz is a real target: `go test -fuzz=Fuzz` runs it.
func Fuzz(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) {
		if decoded, err := Parse(raw); err == nil && decoded != string(raw) {
			t.Fatalf("mismatch for %q", raw)
		}
	})
}
