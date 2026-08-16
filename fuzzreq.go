// Package fuzzreq provides a go/analysis probe for untrusted-input entry
// points with no fuzz target behind them: an exported function taking a bare
// []byte, string, or an io stream and returning a failure is a parser-shaped
// surface, and a package full of them with zero Fuzz targets advertises
// robustness nothing exercises.
//
// This is a PROBE, and it deliberately ships AFTER yze/fuzzassert: mandating
// fuzz targets before a gate can tell an asserting target from an empty one
// invites `f.Fuzz(func(t, b){ _, _ = Parse(b) })`, which satisfies the count
// and verifies nothing.
//
// # What is judged
//
// An exported FREE function declared in a file the go tool compiles into the
// PACKAGE rather than into its test binary, taking at least one untrusted
// parameter and returning at least one result that carries a failure.
//
// A parameter is untrusted when it is a bare string, a bare []byte, or a type
// declared in package io that admits a stream — io.Reader and every interface
// embedding it, io.ReadCloser and io.ReadSeeker among them. A result carries a
// failure when it IMPLEMENTS error, so naming a failure precisely (*ParseError,
// or a defined interface{ Error() string }) buys no exemption over returning
// the bare interface.
//
// Spelling is never vocabulary. An ALIAS is unwrapped — `type Payload = []byte`
// declares no type — and so is a TYPE PARAMETER whose constraint admits exactly
// one type, because `func Parse[T ~[]byte](raw T) error` is []byte at every
// call site and inference leaves every caller unchanged.
//
// A VARIADIC parameter is judged as the SLICE its body sees and not as its
// element, so `raw ...byte` is []byte and reported while `raw ...string` is
// []string and is not. That is a KNOWN HOLE — adding `...` costs one token and
// changes no call site — left open because closing it was measured to cost 293
// findings nobody can act on. See untrustedType, and
// `fuzzreq.variadic-escape-needs-a-discriminator`.
//
// # What is exempt, and why
//
// A named DEFINED type (Name, Payload) is deliberate vocabulary, not an
// untrusted surface — including a package's own stream interface, which is why
// the io clause is keyed on the DECLARING PACKAGE and not on a Reader it can
// find in some scope. A METHOD is out of scope. A type parameter constrained to
// more than one type is a genuine generic rather than a spelling, and there is
// no single type a caller passes. A declaration the go tool compiles only into
// the test binary is out of scope: no importer can reach it, `go test -fuzz`
// cannot fuzz it, and the export_test idiom exists to export it — a finding
// there names a remedy its author cannot take.
//
// # What counts as a fuzz target
//
// The package's targets are read from its DIRECTORY, so internal and external
// test files both count — and only files the BUILD READS count, decided by
// go/build itself: a leading `.` or `_`, a //go:build constraint, and a
// `_GOOS`/`_GOARCH` name suffix all exclude a file, and the `_test.go` suffix
// is matched exactly, as the go tool matches it.
//
// Evaluating a constraint makes the verdict configuration-relative: a target
// compiled only on plan9 discharges nothing on darwin. GOOS and GOARCH agree
// with the driver's, because both read the environment — but BUILD TAGS DO NOT.
// go/build's default context carries no tags, so under `go vet -tags=integration`
// the driver judges a tagged entry point while the evidence reader skips the
// tagged fuzz file beside it, and go/analysis exposes no way to ask which tags
// the driver was given. That disagreement is a defect and it is recorded, not
// papered over: `fuzzreq.evidence-reads-the-tags-the-driver-was-given`. It is
// the same question `errtested` escalated — go/build cannot tell
// `//go:build integration` from `//go:build never_ever_built` — and it is the
// owner's to settle, so it is stated here rather than guessed at.
//
// Within such a file, a target is what `go test` would run: a free function
// named by cmd/go's own isTest rule for the Fuzz prefix, taking exactly one
// *testing.F under whatever local name the file binds package testing to, with
// no results and no type parameters. `Fuzzparse` is not a target, a method is
// not a target, and cmd/go refuses every other signature outright.
//
// A DIRECTORY that cannot be listed counts as fuzzed — the probe demands
// nothing it could not verify the absence of — but a FILE that cannot be read
// or parsed counts as nothing, because that is something an author writes.
package fuzzreq

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// message is the diagnostic for an unfuzzed untrusted-input entry point.
const message = "exported %s consumes untrusted input (%s) and the package has no fuzz target; add a Fuzz target that asserts a property"

// Analyzer reports untrusted-input entry points in fuzz-free packages.
var Analyzer = &analysis.Analyzer{
	Name:     "fuzzreq",
	Doc:      "reports an exported func taking a bare []byte/string or an io stream and returning a failure, in a package with no Fuzz target",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "fuzzreq",
	Categories: []goyze.Category{"tests"},
	URL:        "https://docs.gomatic.dev/yze/fuzzreq",
	Analyzer:   Analyzer,
}

// run reports each qualifying entry point when the package directory holds no
// fuzz target. The scaffolding passes (external test package, test-main) are
// skipped; the plain and in-package-variant passes both judge the production
// declarations and the driver collapses their identical diagnostics.
func run(pass *analysis.Pass) (any, error) {
	if isScaffolding(pass) || len(pass.Files) == 0 {
		return nil, nil
	}
	if hasFuzzTarget(packageDir(pass)) {
		return nil, nil
	}
	ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ins.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, _ := n.(*ast.FuncDecl)
		checkEntryPoint(pass, decl)
	})
	return nil, nil
}

// isScaffolding reports a driver-synthesized test pass: the external test
// package or the test-main package, neither of which carries production
// declarations.
func isScaffolding(pass *analysis.Pass) bool {
	return strings.HasSuffix(pass.Pkg.Name(), "_test") || strings.HasSuffix(pass.Pkg.Path(), ".test")
}

// packageDir is the directory holding the pass's files.
func packageDir(pass *analysis.Pass) dirPath {
	return dirPath(filepath.Dir(pass.Fset.File(pass.Files[0].Pos()).Name()))
}

// checkEntryPoint reports an exported function taking untrusted input and
// returning an error.
func checkEntryPoint(pass *analysis.Pass, decl *ast.FuncDecl) {
	if inTestFile(pass, decl) || decl.Recv != nil || !ast.IsExported(decl.Name.Name) || !returnsError(pass, decl) {
		return
	}
	if input, ok := untrustedParam(pass, decl); ok {
		pass.Reportf(decl.Pos(), message, decl.Name.Name, input)
	}
}

// inTestFile reports a declaration the go tool compiles only into a test
// binary. Such a declaration is exported to the package's own tests and to
// nothing else: no importer can reach it, `go test -fuzz` cannot fuzz it, and
// the export_test idiom exists precisely to export it — so the remedy this
// diagnostic prescribes is not available to its author.
//
// The name comes from the FileSet's own entry rather than from a Position,
// which applies //line directives: a decision ABOUT a file must read something
// that file cannot rewrite. The suffix is compared exactly, because the go
// tool's own rule is exact — `httptest.go`, `a_test.golden.go` and
// `Cased_Test.go` are ordinary compiled source, and each is a fixture.
func inTestFile(pass *analysis.Pass, decl *ast.FuncDecl) bool {
	return strings.HasSuffix(pass.Fset.File(decl.Pos()).Name(), "_test.go")
}

// returnsError reports whether the declaration can report a failure. Any result
// IMPLEMENTING error counts, not only the bare interface: naming a failure
// precisely — `*ParseError`, or a defined `interface{ Error() string }` — is
// better practice than returning `error`, and a rule that reads the spelling
// would hand an exemption to the better-written function.
func returnsError(pass *analysis.Pass, decl *ast.FuncDecl) bool {
	if decl.Type.Results == nil {
		return false
	}
	for _, field := range decl.Type.Results.List {
		if isError(pass.TypesInfo.TypeOf(field.Type)) {
			return true
		}
	}
	return false
}

// isError reports a type that carries a failure.
func isError(t types.Type) bool {
	iface, ok := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
	return ok && t != nil && types.Implements(t, iface)
}
