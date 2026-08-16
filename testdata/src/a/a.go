// Package a holds untrusted-input entry points and NO fuzz target.
package a

import (
	"io"

	"stream"
)

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

// Scan takes a TYPE PARAMETER whose constraint admits exactly one type: []byte
// under a tilde is still []byte at every call site, and inference means no
// caller changes. Making a function generic acquires none of the vocabulary a
// named domain type carries.
func Scan[T ~[]byte](raw T) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}

// ScanText is the same shape one constraint away, over ~string.
func ScanText[T ~string](raw T) error { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}

// ScanTyped is the conforming near-miss: a type parameter whose one term is the
// NAMED domain type, which is vocabulary wherever it is written.
func ScanTyped[T Payload](raw T) error {
	_ = raw
	return nil
}

// Collect takes a VARIADIC string, which is []string and is NOT reported — the
// declared asymmetry with CollectBytes below, and the one place this rule
// deliberately leaves an escape open. Judging the element instead adds 286
// fleet findings on `args ...domain.Argument`, the signature go-app's Runner
// contract requires of every domain Run, whose author can answer only with a
// baseline.
func Collect(raw ...string) error {
	_ = raw
	return nil
}

// CollectBytes takes a variadic byte, which is []byte inside the function and
// was already reported before the element was consulted.
func CollectBytes(raw ...byte) error { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}

// CollectTyped is the conforming near-miss of the variadic: the element is the
// named domain type.
func CollectTyped(raw ...Payload) error {
	_ = raw
	return nil
}

// Close takes an io.ReadCloser, which IS an io.Reader — the interface embeds
// it, every caller passes the same untrusted stream, and widening the parameter
// from Reader to ReadCloser is one word.
func Close(rc io.ReadCloser) error { // want `consumes untrusted input \(io.ReadCloser\)`
	return rc.Close()
}

// Seek takes an io.ReadSeeker, the same shape through a different embedding.
func Seek(rs io.ReadSeeker) error { // want `consumes untrusted input \(io.ReadSeeker\)`
	_, err := io.ReadAll(rs)
	return err
}

// Sink takes an io.Writer, which is the conforming near-miss: an interface from
// the same package that reads nothing.
func Sink(w io.Writer) error {
	_, err := w.Write(nil)
	return err
}

// Fault is a defined error type of this package's own.
type Fault struct{ reason string }

// Error makes Fault an error.
func (f Fault) Error() string { return f.reason }

// Detect returns a CONCRETE error type. It can fail, which is the whole of what
// the result clause asks; naming the failure precisely is better practice than
// returning the bare interface, and it must not buy an exemption.
func Detect(raw []byte) *Fault { // want `consumes untrusted input \(\[\]byte\)`
	_ = raw
	return nil
}

// Classify returns a DEFINED interface whose method set is error's.
type failure interface{ Error() string }

// Classify returns that interface.
func Classify(raw string) failure { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}

// Reject is the conforming near-miss of the result clause: a defined type with
// no Error method at all cannot report a failure.
type Reject struct{ reason string }

// Describe returns a type that is not an error.
func Describe(raw []byte) Reject {
	_ = raw
	return Reject{reason: string(raw)}
}

// ScanEither is a genuine generic rather than a spelling: its constraint admits
// TWO types, so no single type is what a caller passes and there is nothing to
// resolve it to. One term is the deviation from Scan above.
func ScanEither[T ~[]byte | ~string](raw T) error {
	_ = raw
	return nil
}

// Body is this package's OWN stream interface — structurally io.Reader, and
// deliberate vocabulary all the same, because it is declared here.
type Body interface {
	Read(p []byte) (int, error)
}

// Consume takes the package's own interface: the conforming near-miss of the io
// clause, differing from Close only in which package declares the type.
func Consume(b Body) error {
	_, err := io.ReadAll(b)
	return err
}

// Siphon2 takes another package's Body — a named domain type, in a package that
// happens to declare a `Reader` of its own. It differs from Close in exactly
// one place: which package declares the type.
func Siphon2(b stream.Body) error {
	_, err := io.ReadAll(b)
	return err
}

// Wrap takes the universe's own `error`, whose type object belongs to NO
// package. It is the shape that reaches the io clause with a nil package, and
// without the guard there the analyzer dereferences nil and takes the whole run
// down — every other rule in the suite with it.
func Wrap(e error) error {
	return e
}

// Anything is constrained by `any`, whose constraint interface embeds no term
// at all: there is nothing for a caller to be passing, so nothing is resolved.
func Anything[T any](raw T) error {
	_ = raw
	return nil
}
