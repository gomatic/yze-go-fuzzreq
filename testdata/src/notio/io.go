// Package io is NAMED io and is not io: its import path is notio. It deviates
// from the real package in exactly ONE dimension, the path, which is the
// dimension the clause is keyed on — a fixture differing in name AND path
// cannot tell the two apart.
package io

// Reader is this package's own stream interface, spelled exactly as io's.
type Reader interface {
	Read(p []byte) (int, error)
}

// Body is a domain type built on it.
type Body interface {
	Reader
	Close() error
}
