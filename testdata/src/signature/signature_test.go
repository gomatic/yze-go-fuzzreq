package signature

import "testing"

// Fuzzparse deviates from bare's `Fuzz` in exactly one place — the rune after
// the prefix is lower case — and in nothing else: it is free, it takes one
// *testing.F, it compiles, and it reads like a fuzz target to anyone scanning
// the file. cmd/go's isTest rule says it is not one, so `go test -list
// 'Fuzz.*'` prints nothing and `go test -fuzz=Fuzzparse` matches nothing. One
// keystroke off FuzzParse, and the package leaves the rule.
func Fuzzparse(f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) {
		if decoded, err := Parse(raw); err == nil && decoded != string(raw) {
			t.Fatalf("mismatch for %q", raw)
		}
	})
}

// helper carries a method wearing the marker.
type helper struct{}

// FuzzMethod deviates in exactly one place from a target too: it has a
// RECEIVER. It compiles, `go test -list 'Fuzz.*'` lists nothing for it, and
// there is no way to ask the go tool to fuzz a method.
func (helper) FuzzMethod(f *testing.F) { _ = f }

// Fuzzy is the shape the intent review reproduced. It deviates in TWO places —
// the name and the signature — so it pins neither boundary on its own and is
// kept only because it is what a real evasion looked like.
func Fuzzy(n int) int { return n }
