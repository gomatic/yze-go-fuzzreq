package fuzzreq

import (
	"go/ast"
	"go/token"
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// errStub reports the stubbed directory failure.
const errStub errs.Const = "stubbed"

// TestHasFuzzTargetFailsOpen pins the probe's fail-open contract: a directory
// it cannot list, and a test file it cannot parse, both count as fuzzed —
// the probe demands nothing it could not verify the absence of.
func TestHasFuzzTargetFailsOpen(t *testing.T) {
	original := readDir
	t.Cleanup(func() { readDir = original })

	readDir = func(dirPath) ([]string, error) { return nil, errStub }
	assert.True(t, hasFuzzTarget("anywhere"))

	readDir = func(dirPath) ([]string, error) { return []string{"broken_test.go"}, nil }
	assert.True(t, hasFuzzTarget("/nonexistent-dir"), "an unparseable test file fails open")
}

// TestOSReadDirNames pins the real directory lister, including its error.
func TestOSReadDirNames(t *testing.T) {
	t.Parallel()

	_, err := osReadDirNames("/nonexistent-directory-for-fuzzreq")
	assert.Error(t, err)
}

// declAt is a function declaration positioned in one file of a FileSet.
func declAt(at token.Pos) *ast.FuncDecl {
	return &ast.FuncDecl{Name: ast.NewIdent("Entry"), Type: &ast.FuncType{Func: at, Params: &ast.FieldList{}}}
}

// TestInTestFileComparesTheSuffixExactly pins the exemption to the go tool's
// own rule, which is an exact `_test.go` suffix on the name the FileSet holds.
// Each name below deviates from `entry_test.go` in exactly ONE dimension the
// matcher could be widened along — the underscore, the position of the literal,
// letter case, and the name the file claims for itself — and every one of them
// is a file the go tool compiles as ordinary source. A widening therefore
// silences a real entry point, which is a disablement bought with a filename.
func TestInTestFileComparesTheSuffixExactly(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	exempt := map[string]bool{
		"entry_test.go":        true,
		"httptest.go":          false,
		"entry_test.golden.go": false,
		"Entry_Test.go":        false,
		"entry_TEST.go":        false,
	}
	for name, want := range exempt {
		file := fset.AddFile(name, -1, 100)
		assert.Equal(t, want, inTestFile(&analysis.Pass{Fset: fset}, declAt(file.Pos(1))), name)
	}

	claimant := fset.AddFile("compiled.go", -1, 100)
	claimant.AddLineColumnInfo(0, "zz_test.go", 1, 1)
	assert.Equal(t, "zz_test.go", fset.Position(claimant.Pos(1)).Filename,
		"the position machinery does adopt the claimed name — without this the assertion below holds vacuously")
	assert.False(t, inTestFile(&analysis.Pass{Fset: fset}, declAt(claimant.Pos(1))),
		"a //line directive is what AddLineColumnInfo compiles to, and it renames nothing the go tool reads")
}
