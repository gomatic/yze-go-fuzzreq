package a

import "testing"

// TestParse is an ordinary test — NOT a fuzz target, so the package still
// owes its entry points fuzz coverage.
func TestParse(t *testing.T) {
	if decoded, err := Parse([]byte("x")); err != nil || decoded != "x" {
		t.Fatalf("Parse = %q, %v", decoded, err)
	}
}
