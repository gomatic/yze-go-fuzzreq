// Package a holds untrusted-input entry points and NO fuzz target.
package a

import "io"

// Payload is a named domain type — deliberate vocabulary, not bare input.
type Payload []byte

// Parse consumes bare bytes and can fail: the fuzz-shaped surface.
func Parse(raw []byte) (string, error) { // want `consumes untrusted input \(\[\]byte\)`
	return string(raw), nil
}

// Decode consumes a bare string.
func Decode(raw string) error { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}

// Drain consumes an io.Reader.
func Drain(r io.Reader) error { // want `consumes untrusted input \(io.Reader\)`
	_, err := io.ReadAll(r)
	return err
}

// ParseTyped consumes the NAMED type; vocabulary, not bare input.
func ParseTyped(raw Payload) (string, error) {
	return string(raw), nil
}

// Render cannot fail, so it is no parser surface.
func Render(raw []byte) string { return string(raw) }

// Emit returns nothing, so it is no parser surface either.
func Emit(raw []byte) { _ = raw }

// parse is unexported: not an entry point.
func parse(raw []byte) error {
	_ = raw
	return nil
}
