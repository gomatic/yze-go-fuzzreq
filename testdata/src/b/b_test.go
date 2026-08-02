package b

import "testing"

// FuzzParse asserts the parse surface.
func FuzzParse(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) {
		decoded, err := Parse(raw)
		if err == nil && decoded != string(raw) {
			t.Fatalf("mismatch for %q", raw)
		}
	})
}
