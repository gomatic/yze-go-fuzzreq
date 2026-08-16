// Package aliased is fuzzed through an ALIASED import of testing, which the go
// tool runs exactly like any other target — `go test -list 'Fuzz.*'` prints
// FuzzAliased. A matcher hard-coding the spelling `testing` would report this
// package, which is a finding whose author has already done what it asks.
package aliased

// Parse consumes bare bytes and can fail, and is fuzzed, so nothing is owed.
func Parse(raw []byte) (string, error) {
	return string(raw), nil
}
