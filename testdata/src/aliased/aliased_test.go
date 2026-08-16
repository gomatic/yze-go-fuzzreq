package aliased

import tst "testing"

// FuzzAliased is a real target reached through the alias.
func FuzzAliased(f *tst.F) {
	f.Fuzz(func(t *tst.T, raw []byte) {
		if decoded, err := Parse(raw); err == nil && decoded != string(raw) {
			t.Fatalf("mismatch for %q", raw)
		}
	})
}
