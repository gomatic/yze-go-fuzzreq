package dot

import . "testing"

// A dot import brings F into scope unqualified, and `go test -list 'Fuzz.*'`
// runs a target declared this way exactly like any other. Missing it would
// report a package that IS fuzzed, which is a finding its author cannot act on.

// FuzzDot is a real target named through a dot import.
func FuzzDot(f *F) { _ = f }
