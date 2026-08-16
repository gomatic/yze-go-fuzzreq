package wrongsig

import "testing"

// This file is NOT under testdata/src and no package loader ever reads it. That
// is deliberate: every declaration below is a Fuzz-NAMED free function that
// cmd/go refuses, so a package carrying one does not build and analysistest
// cannot load it —
//
//	wrong signature for FuzzBad, must be: func FuzzBad(f *testing.F)
//	wrong signature for FuzzGeneric, test functions cannot have type parameters
//
// The corpus therefore cannot case these shapes at all, and TestDeclaresFuzz
// pins them against declaresFuzz directly, which is the seam that reads a file
// from disk rather than from a loaded package.

// FuzzBad is what a prefix-only matcher credits: a package exempted by a
// declaration that makes the package untestable.
func FuzzBad(n int) int { return n }

// FuzzTwo has the right first parameter and one too many.
func FuzzTwo(f *testing.F, n int) { _, _ = f, n }

// FuzzRes has the right parameter and a result.
func FuzzRes(f *testing.F) error { _ = f; return nil }

// FuzzGeneric is refused by name.
func FuzzGeneric[T any](f *testing.F) { _ = f }

// FuzzWrongType is the near-miss of isTestingF: free, Fuzz-named, exactly one
// parameter, and the parameter is the wrong type.
func FuzzWrongType(t *testing.T) { _ = t }

// FuzzPair is the near-miss of the parameter COUNT read through one field:
// `f, g *testing.F` is one field holding two names.
func FuzzPair(f, g *testing.F) { _, _ = f, g }

// FuzzValue is the near-miss of the POINTER: the type is testing.F itself, not
// *testing.F, which cmd/go refuses like every other signature here.
func FuzzValue(f testing.F) { _ = f }

// FuzzBareF names F unqualified in a file that imports testing normally, which
// is the near-miss of the QUALIFIER — only a dot import puts F in scope.
func FuzzBareF(f *F) { _ = f }
