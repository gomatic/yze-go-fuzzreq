package orphan

// This file names *testing.F without importing testing, so it cannot compile
// and the go tool would never run anything in it. It is the near-miss of
// testingName: the spelling is right and the binding is absent.

// FuzzOrphan wears the marker in a file that names no package testing.
func FuzzOrphan(f *testing.F) { _ = f }
