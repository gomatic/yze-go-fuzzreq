//line zz_test.go:1
package a

// The line above is the entire difference between this file and a.go, and it
// is a compiler feature for generated code: the go tool still compiles
// linetest.go as ordinary source and `go list` still names it in GoFiles. Only
// fset.Position adopts the claimed name, so an exemption resolved through
// Position is decided by the very file being judged — one comment line and a
// real entry point leaves the rule, with nothing in any configuration file to
// find. The FileSet's own entry is what a decision about a file must read.

// Rename consumes bare bytes from a file claiming to be a test.
func Rename(raw []byte) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}
