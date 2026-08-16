package a

import "testing"

// TestParse is an ordinary test — NOT a fuzz target, so the package still
// owes its entry points fuzz coverage.
func TestParse(t *testing.T) {
	if decoded, err := Parse([]byte("x")); err != nil || decoded != "x" {
		t.Fatalf("Parse = %q, %v", decoded, err)
	}
}

// Helper is the export_test idiom: exported from the in-package test variant so
// a sibling test can reach an unexported symbol. It is not part of the
// package's surface — no importer can call it, `go test -fuzz` cannot fuzz it,
// and unexporting a file whose whole purpose is to export is not a remedy. The
// rule judges PRODUCTION declarations, so nothing is reported here.
func Helper(raw []byte) error {
	_ = raw
	return nil
}
