package plain

import "testing"

// FuzzPlain is the ordinary shape, here as the positive control: without it a
// green result from TestDeclaresFuzz could mean the reader never found a target
// at all.
func FuzzPlain(f *testing.F) { _ = f }
