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

// Text is an ALIAS of string: one `=` away from Payload, and the whole
// difference. An alias declares no type — types.Identical(Text, string) holds —
// so it carries none of the vocabulary the exemption exists for.
type Text = string

// Blob is an alias of []byte.
type Blob = []byte

// Stream is an alias of io.Reader.
type Stream = io.Reader

// Mirror is an alias OF the named domain type, which is the conforming side of
// the same boundary: the type it spells is still Payload.
type Mirror = Payload

// Octet is an alias of byte, which puts the same spelling one level down, in
// the slice ELEMENT rather than in the parameter.
type Octet = byte

// Announce consumes a bare string under an alias spelling.
func Announce(raw Text) error { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}

// Absorb consumes bare bytes under an alias spelling.
func Absorb(raw Blob) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}

// Siphon consumes an io.Reader under an alias spelling.
func Siphon(r Stream) error { // want `consumes untrusted input \(io.Reader\)`
	_, err := io.ReadAll(r)
	return err
}

// ReflectTyped consumes the alias OF the named type: still vocabulary.
func ReflectTyped(raw Mirror) error {
	_ = raw
	return nil
}

// Ingest consumes a slice whose ELEMENT is the alias spelling of byte.
func Ingest(raw []Octet) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
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
