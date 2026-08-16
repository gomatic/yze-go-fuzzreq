package a

// This file's name CONTAINS `_test.go` and does not end in it, so the go tool
// compiles it as ordinary source and `go list` puts it in GoFiles. Widening the
// exemption from a suffix to a substring would silence the entry point below.

// Golden consumes a bare string from a file whose name contains the literal.
func Golden(raw string) error { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}
