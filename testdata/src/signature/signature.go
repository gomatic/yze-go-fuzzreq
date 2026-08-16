// Package signature holds an entry point beside every Fuzz-NAMED declaration
// the go tool would refuse to run. None of them is a fuzz target, so the entry
// point is owed one exactly as if the file were not there.
package signature

// Parse consumes bare bytes and can fail.
func Parse(raw []byte) (string, error) { // want `consumes untrusted input \(\[\]byte\)`
	return string(raw), nil
}
