// Package fuzzreq provides a go/analysis probe for untrusted-input entry
// points with no fuzz target behind them: an exported function taking a bare
// []byte, string, or io.Reader and returning an error is a parser-shaped
// surface, and a package full of them with zero Fuzz targets advertises
// robustness nothing exercises.
//
// This is a PROBE, and it deliberately ships AFTER yze/fuzzassert: mandating
// fuzz targets before a gate can tell an asserting target from an empty one
// invites `f.Fuzz(func(t, b){ _, _ = Parse(b) })`, which satisfies the count
// and verifies nothing. Bare types only: a named domain type (Name, Payload)
// is deliberate vocabulary, not an untrusted byte surface, and methods are
// out of scope. The package's fuzz targets are read from its DIRECTORY, so
// internal and external test files both count.
package fuzzreq

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// message is the diagnostic for an unfuzzed untrusted-input entry point.
const message = "exported %s consumes untrusted input (%s) and the package has no fuzz target; add a Fuzz target that asserts a property"

// dirPath is a package's directory on disk.
type dirPath string

// readDir is the seam tests replace; production lists the package directory.
var readDir = osReadDirNames

// Analyzer reports untrusted-input entry points in fuzz-free packages.
var Analyzer = &analysis.Analyzer{
	Name:     "fuzzreq",
	Doc:      "reports an exported func taking []byte/string/io.Reader and returning error, in a package with no Fuzz target",
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

// hasFuzzTarget reports whether any test file in dir declares a Fuzz
// function. An unreadable directory or file counts as having one: the probe
// fails open rather than demand fuzzing it cannot verify the absence of.
func hasFuzzTarget(dir dirPath) bool {
	names, err := readDir(dir)
	if err != nil {
		return true
	}
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") && declaresFuzz(testFilePath(filepath.Join(string(dir), name))) {
			return true
		}
	}
	return false
}

// testFilePath locates one test file on disk.
type testFilePath string

// declaresFuzz reports whether the file declares a top-level Fuzz* function.
func declaresFuzz(path testFilePath) bool {
	parsed, err := parser.ParseFile(token.NewFileSet(), string(path), nil, parser.SkipObjectResolution)
	if err != nil {
		return true
	}
	for _, decl := range parsed.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Fuzz") {
			return true
		}
	}
	return false
}

// checkEntryPoint reports an exported function taking untrusted input and
// returning an error.
func checkEntryPoint(pass *analysis.Pass, decl *ast.FuncDecl) {
	if decl.Recv != nil || !ast.IsExported(decl.Name.Name) || !returnsError(pass, decl) {
		return
	}
	if input, ok := untrustedParam(pass, decl); ok {
		pass.Reportf(decl.Pos(), message, decl.Name.Name, input)
	}
}

// returnsError reports whether the declaration's results include error.
func returnsError(pass *analysis.Pass, decl *ast.FuncDecl) bool {
	if decl.Type.Results == nil {
		return false
	}
	for _, field := range decl.Type.Results.List {
		if types.Identical(pass.TypesInfo.TypeOf(field.Type), types.Universe.Lookup("error").Type()) {
			return true
		}
	}
	return false
}

// untrustedName describes which untrusted shape a parameter carries.
type untrustedName string

// untrustedParam reports the first parameter carrying a BARE untrusted type:
// []byte, string, or io.Reader. A named domain type is deliberate vocabulary
// and does not count.
func untrustedParam(pass *analysis.Pass, decl *ast.FuncDecl) (untrustedName, bool) {
	for _, field := range decl.Type.Params.List {
		if input, ok := untrustedType(pass.TypesInfo.TypeOf(field.Type)); ok {
			return input, true
		}
	}
	return "", false
}

// untrustedType classifies a bare untrusted parameter type.
func untrustedType(t types.Type) (untrustedName, bool) {
	switch resolved := t.(type) {
	case *types.Basic:
		if resolved.Kind() == types.String {
			return "string", true
		}
	case *types.Slice:
		if isByte(resolved.Elem()) {
			return "[]byte", true
		}
	case *types.Named:
		if isIOReader(resolved) {
			return "io.Reader", true
		}
	}
	return "", false
}

// isByte reports the byte element type.
func isByte(t types.Type) bool {
	basic, ok := t.(*types.Basic)
	return ok && basic.Kind() == types.Byte
}

// isIOReader reports io's Reader interface by declaring path and name.
func isIOReader(named *types.Named) bool {
	return named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "io" && named.Obj().Name() == "Reader"
}

// osReadDirNames lists the entry names of a directory.
func osReadDirNames(dir dirPath) ([]string, error) {
	entries, err := os.ReadDir(string(dir))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}
