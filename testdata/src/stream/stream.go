// Package stream declares a Reader of its OWN, which is what most packages that
// talk about streams do. Nothing here is io, and nothing here is bare.
package stream

// Reader is this package's stream interface, structurally identical to
// io.Reader and named the same — which is exactly why the io clause must be
// keyed on the DECLARING PACKAGE and not on a name it can find in any scope.
type Reader interface {
	Read(p []byte) (int, error)
}

// Body is a domain type built on it.
type Body interface {
	Reader
	Close() error
}
