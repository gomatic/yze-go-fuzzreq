package brokenbody

import "testing"

// The HEADER parses, so go/build includes the file; the BODY does not, so the
// marker reader cannot see what it declares. Crediting it would exempt a
// package for a file `go test` refuses to compile.
func FuzzBroken(f *testing.F) { this is not go
