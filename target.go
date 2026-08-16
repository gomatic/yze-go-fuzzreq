package fuzzreq

import (
	"go/ast"
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
		return true
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
