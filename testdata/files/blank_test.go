package blank

import _ "testing"

// A blank import binds no name, so `_` is not a package qualifier and this file
// does not compile. It parses, though, which is all the marker reader sees —
// so without testingName returning the empty name it would be credited.

// FuzzBlank wears the marker through an import that names nothing.
func FuzzBlank(f *_.F) { _ = f }
