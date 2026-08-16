package fuzzreq

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// dirPath is a package's directory on disk.
type dirPath string

// readDir is the seam tests replace; production lists the package directory.
var readDir = osReadDirNames

// buildContext decides which files the go tool reads, and is injected so a case
// can pin the verdict on a platform other than the one the test runs on. It is
// build.Default rather than a hand-built context because the DRIVER already
// filtered pass.Files for exactly this configuration: the entry points judged
// and the evidence credited must come from one build, or a package is asked to
// fuzz code the same build excluded.
var buildContext = build.Default

// hasFuzzTarget reports whether any test file in dir declares a Fuzz function.
// An unreadable DIRECTORY counts as having one — the probe demands nothing it
// could not verify the absence of — but a file the go tool would not read never
// counts, whatever it declares.
func hasFuzzTarget(dir dirPath) bool {
	names, err := readDir(dir)
	if err != nil {
		return true
	}
	for _, name := range names {
		if compiledTestFile(dir, fileName(name)) && declaresFuzz(testFilePath(filepath.Join(string(dir), name))) {
			return true
		}
	}
	return false
}

// fileName is one entry of a package directory.
type fileName string

// compiledTestFile reports a file the go tool compiles INTO THE TEST BINARY: an
// exact `_test.go` suffix, and inclusion decided by go/build itself, which
// applies the rules the compiler applies — a leading `.` or `_` is skipped, a
// `//go:build` constraint is evaluated, and a `_GOOS`/`_GOARCH` name suffix is
// honoured. Evidence that the build never reads is no evidence: `go test`
// prints "[no test files]" for a package whose only target is one of those, so
// crediting it exempts a package that fuzzes nothing, with nothing written down
// anywhere for an inventory to find.
func compiledTestFile(dir dirPath, name fileName) bool {
	if !strings.HasSuffix(string(name), "_test.go") {
		return false
	}
	included, err := buildContext.MatchFile(string(dir), string(name))
	return err == nil && included
}

// testFilePath locates one test file on disk.
type testFilePath string

// declaresFuzz reports whether the file declares a fuzz target THE GO TOOL
// WOULD RUN: a free function whose name passes cmd/go's own isTest rule for the
// Fuzz prefix, taking exactly one *testing.F and returning nothing. Every other
// spelling is either refused by `go test` outright — "wrong signature for
// FuzzBad, must be: func FuzzBad(f *testing.F)" — or silently never run, as a
// method or a lower-cased `Fuzzy` is, so crediting one would exempt a package
// for a marker that fuzzes nothing.
func declaresFuzz(path testFilePath) bool {
	parsed, err := parser.ParseFile(token.NewFileSet(), string(path), nil, parser.SkipObjectResolution)
	if err != nil {
		return false
	}
	local := testingName(parsed)
	for _, decl := range parsed.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && isFuzzTarget(fn, local) {
			return true
		}
	}
	return false
}

// isFuzzTarget reports a declaration the go tool would run as a fuzz target: a
// FREE function (a method is unreachable by `go test -fuzz`), named by cmd/go's
// isTest rule, with the one signature cmd/go accepts.
func isFuzzTarget(fn *ast.FuncDecl, local testingLocal) bool {
	return fn.Recv == nil && isFuzzName(funcName(fn.Name.Name)) && takesTestingF(fn, local)
}

// testingLocal is the name a file binds package testing to.
type testingLocal string

// testingName is the local name package testing is bound to in the file, and
// the empty name when the file cannot name it at all — it does not import
// testing, or imports it blank, which binds nothing. No identifier is spelled
// "", so an unnameable type credits nothing without a second guard saying so.
func testingName(file *ast.File) testingLocal {
	for _, spec := range file.Imports {
		if spec.Path.Value != `"testing"` {
			continue
		}
		if spec.Name == nil {
			return "testing"
		}
		if spec.Name.Name == "_" {
			return ""
		}
		return testingLocal(spec.Name.Name)
	}
	return ""
}

// funcName is a declared function's identifier.
type funcName string

// isFuzzName mirrors cmd/go/internal/load.isTest: the prefix, then either the
// end of the name or a rune that is not lower case. It is why `Fuzzy` is
// ordinary code and `Fuzz` is a target.
func isFuzzName(name funcName) bool {
	const prefix = "Fuzz"
	if !strings.HasPrefix(string(name), prefix) {
		return false
	}
	if len(name) == len(prefix) {
		return true
	}
	first, _ := utf8.DecodeRuneInString(string(name[len(prefix):]))
	return !unicode.IsLower(first)
}

// takesTestingF reports the exact signature `func(*testing.F)`: one parameter,
// one name at most, no results and no type parameters. `go test` rejects every
// other shape of a Fuzz-named function rather than running it.
func takesTestingF(fn *ast.FuncDecl, local testingLocal) bool {
	if fn.Type.Results != nil || fn.Type.TypeParams != nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	param := fn.Type.Params.List[0]
	return len(param.Names) <= 1 && isTestingF(param.Type, local)
}

// isTestingF reports the type expression *<local>.F, honouring the name the
// file binds package testing to: `import tst "testing"` declares targets the go
// tool runs, so a matcher hard-coding the spelling would report a fuzzed
// package.
func isTestingF(expr ast.Expr, local testingLocal) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	if local == "." {
		ident, isIdent := star.X.(*ast.Ident)
		return isIdent && ident.Name == "F"
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "F" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == string(local)
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
