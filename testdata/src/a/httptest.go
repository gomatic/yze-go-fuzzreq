package a

// This file is ORDINARY COMPILED SOURCE. `go list` puts httptest.go in GoFiles
// — the go tool's test-file rule is the exact suffix `_test.go`, and
// `test.go` is not it. Widening the exemption to a `test.go` suffix would
// silence a real entry point, which is the left edge of the same literal.
// net/http/httptest/httptest.go is the shape wearing this name in the wild.

// Serve consumes bare bytes from a file named for a test helper.
func Serve(raw []byte) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}
