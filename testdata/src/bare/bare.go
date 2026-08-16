// Package bare sits at the other side of the name boundary: its target is
// named exactly `Fuzz`, the shortest name the go tool's isTest rule accepts —
// one character from `Fuzzy`, which it rejects.
package bare

// Parse consumes bare bytes and can fail, and is fuzzed, so nothing is owed.
func Parse(raw []byte) (string, error) {
	return string(raw), nil
}
