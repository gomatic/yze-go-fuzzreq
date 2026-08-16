package aliased

import tst "testing"

// FuzzAliased is a real target reached through an aliased import — `go test
// -list 'Fuzz.*'` prints it. Credited here as well as in testdata/src/aliased,
// because this is the seam where the local name is resolved.
func FuzzAliased(f *tst.F) { _ = f }
