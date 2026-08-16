// Package unbuilt holds one entry point beside every fuzz target the GO TOOL
// NEVER READS. `go test ./...` prints "[no test files]" for a package whose
// only targets are these, so each is a marker acquired without the property —
// and unlike a configuration entry, none of them appears in any file an
// inventory knows to open.
package unbuilt

// Parse consumes bare bytes and can fail.
func Parse(raw []byte) (string, error) { // want `consumes untrusted input \(\[\]byte\)`
	return string(raw), nil
}
