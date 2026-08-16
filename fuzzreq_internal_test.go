package fuzzreq

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"path/filepath"
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// errStub reports the stubbed directory failure.
const errStub errs.Const = "stubbed"

// TestHasFuzzTargetFailsOpenForTheDirectoryAndClosedForAFile pins where the
// fail-open contract ends. A directory the probe cannot LIST is a state it
// cannot reason about, so it demands nothing. A single FILE it cannot read or
// parse is different in kind: it is something an author writes, and crediting
// it hands out the exemption for a file the go tool refuses to compile.
//
// The test this replaces asserted the opposite ("an unparseable test file fails
// open") and did not even reach that branch — it stubbed readDir to name a file
// under a directory that does not exist, so what it entered was file-not-found.
// Both branches are entered here, and named.
func TestHasFuzzTargetFailsOpenForTheDirectoryAndClosedForAFile(t *testing.T) {
	original := readDir
	t.Cleanup(func() { readDir = original })

	readDir = func(dirPath) ([]string, error) { return nil, errStub }
	assert.True(t, hasFuzzTarget("anywhere"), "a directory that cannot be listed demands nothing")

	readDir = func(dirPath) ([]string, error) { return []string{"absent_test.go"}, nil }
	assert.False(t, hasFuzzTarget("/nonexistent-dir"), "a file that cannot be read is not evidence")

	files := dirPath(filepath.Join("testdata", "files"))
	readDir = func(dirPath) ([]string, error) { return []string{"brokenbody_test.go"}, nil }
	assert.False(t, hasFuzzTarget(files), "a file that cannot be parsed is not evidence")

	readDir = func(dirPath) ([]string, error) { return []string{"plain_test.go"}, nil }
	assert.True(t, hasFuzzTarget(files), "the positive control: a real target in a readable file")
}

// TestOSReadDirNames pins the real directory lister, including WHICH error it
// produces. The contract is that the os error reaches the caller unwrapped, so
// hasFuzzTarget's fail-open branch is entered by a directory that is genuinely
// absent and not by anything else; asserting only that "an error occurred"
// holds for any failure whatever and so tests nothing about that contract.
func TestOSReadDirNames(t *testing.T) {
	t.Parallel()

	names, err := osReadDirNames("/nonexistent-directory-for-fuzzreq")
	assert.ErrorIs(t, err, fs.ErrNotExist)
	assert.Nil(t, names, "no names come back with the error")

	names, err = osReadDirNames(dirPath(filepath.Join("testdata", "files")))
	assert.NoError(t, err, "the positive control: the assertion above is about the directory, not the lister")
	assert.Contains(t, names, "plain_test.go")
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

// TestDeclaresFuzz pins the marker to the signature the go tool requires,
// against files no package loader can read. That is not a convenience: cmd/go
// REFUSES a Fuzz-named free function of any other signature — "wrong signature
// for FuzzBad, must be: func FuzzBad(f *testing.F)" — so a package carrying one
// does not load, and analysistest cannot express the case at all. The corpus
// covers the two forgeries a package can carry and still build (a lower-cased
// name and a method); these are the rest.
func TestDeclaresFuzz(t *testing.T) {
	t.Parallel()

	credited := map[string]bool{
		"plain_test.go":      true,
		"aliased_test.go":    true,
		"dot_test.go":        true,
		"brokenbody_test.go": false,
		"wrongsig_test.go":   false,
		"orphan_test.go":     false,
		"blank_test.go":      false,
	}
	for name, want := range credited {
		assert.Equal(t, want, declaresFuzz(testFilePath(filepath.Join("testdata", "files", name))), name)
	}
}

// TestTestingNameIsTheNameTheFileCanWrite pins the invariant testingName
// documents: the empty name is returned exactly when the file CANNOT write the
// type, so no second guard is needed to refuse a target it names anyway. A
// blank import binds nothing and a missing import binds nothing; both are the
// empty name, and no identifier is spelled "".
func TestTestingNameIsTheNameTheFileCanWrite(t *testing.T) {
	t.Parallel()

	bindings := map[string]testingLocal{
		`package p; import "testing"`:      "testing",
		`package p; import tst "testing"`:  "tst",
		`package p; import . "testing"`:    ".",
		`package p; import _ "testing"`:    "",
		`package p`:                        "",
		`package p; import "text/testing"`: "",
	}
	for src, want := range bindings {
		parsed, err := parser.ParseFile(token.NewFileSet(), "p.go", src, parser.SkipObjectResolution)
		assert.NoError(t, err, src)
		assert.Equal(t, want, testingName(parsed), src)
	}
}

// TestCompiledTestFileReadsTheGoToolsOwnRules pins the evidence side to the
// files the build actually reads. Every name below sits in testdata/src/unbuilt
// beside a real, correctly-signed fuzz target, and `go test ./...` prints "[no
// test files]" for that package — so each is a marker acquired without the
// property, and none of them leaves an entry in any configuration file.
func TestCompiledTestFileReadsTheGoToolsOwnRules(t *testing.T) {
	t.Parallel()

	dir := dirPath(filepath.Join("testdata", "src", "unbuilt"))
	compiled := map[fileName]bool{
		".silence_test.go":      false,
		"_silence_test.go":      false,
		"tagged_test.go":        false,
		"unbuilt_plan9_test.go": false,
		"Cased_Test.go":         false,
		"fuzztarget.go":         false,
		"helpertest.go":         false,
		"zz_test.golden.go":     false,
		"absent_test.go":        false,
	}
	for name, want := range compiled {
		assert.Equal(t, want, compiledTestFile(dir, name), string(name))
	}
	assert.True(t, compiledTestFile(dirPath(filepath.Join("testdata", "files")), "plain_test.go"),
		"the positive control: without it every false above could mean the reader never worked")
}

// TestCompiledTestFileIsDecidedByTheBuildContext pins the verdict on a second
// platform, which docs/s03.md requires of anything platform-dependent — and
// this rule is: a constraint is evaluated, so a target compiled only on plan9
// discharges nothing on darwin and everything on plan9. That is not a defect to
// remove but a fact to state, because the DRIVER already filtered the entry
// points the same way; evidence and findings must come from one build.
func TestCompiledTestFileIsDecidedByTheBuildContext(t *testing.T) {
	original := buildContext
	t.Cleanup(func() { buildContext = original })

	dir := dirPath(filepath.Join("testdata", "src", "unbuilt"))
	buildContext.GOOS = "darwin"
	assert.False(t, compiledTestFile(dir, "unbuilt_plan9_test.go"), "darwin never compiles it")
	buildContext.GOOS = "plan9"
	assert.True(t, compiledTestFile(dir, "unbuilt_plan9_test.go"), "plan9 always does")
}

// TestIOStreamNameNeedsIOsOwnReader pins the guard that keeps the io clause
// from dereferencing nothing. The lookup is what decides "admits a stream", and
// it is scoped to the DECLARING package — so a package answering to io's path
// while declaring no Reader is the one input that reaches it. No source file
// can produce that (the real io always has Reader, and any other package has a
// different path), so the case is built rather than parsed.
func TestIOStreamNameNeedsIOsOwnReader(t *testing.T) {
	t.Parallel()

	pkg := types.NewPackage("io", "io")
	obj := types.NewTypeName(token.NoPos, pkg, "Stream", nil)
	named := types.NewNamed(obj, types.NewInterfaceType(nil, nil), nil)
	pkg.MarkComplete()

	name, ok := ioStreamName(named)
	assert.False(t, ok, "no Reader in scope, so nothing is known to admit a stream")
	assert.Empty(t, name)
}

// TestUntrustedTypeJudgesAVariadicAsTheSliceItIs pins WHERE THE RULE IS TODAY,
// at the type level where a variadic parameter actually arrives: `raw ...byte`
// is []byte and reported, `raw ...string` is []string and is not.
//
// The second assertion pins a HOLE, not a contract, and says so where the next
// reader will look: adding `...` to a string parameter costs one token and
// changes no call site. It is asserted because closing it was measured to cost
// 293 fleet findings nobody can act on, so the silence must not change by
// accident before a discriminator exists — `fuzzreq.variadic-escape-needs-a-discriminator`.
// A test asserting a limitation is a defect until the limitation is closed; this
// one is recorded as such rather than described as a design.
func TestUntrustedTypeJudgesAVariadicAsTheSliceItIs(t *testing.T) {
	t.Parallel()

	name, ok := untrustedType(types.NewSlice(types.Typ[types.Byte]))
	assert.True(t, ok, "the slice a `...byte` parameter is")
	assert.Equal(t, untrustedName("[]byte"), name)

	name, ok = untrustedType(types.NewSlice(types.Typ[types.String]))
	assert.False(t, ok, "the slice a `...string` parameter is — the open hole, not a contract")
	assert.Empty(t, name)
}

// TestSoleEmbeddedNeedsExactlyOneEmbeddingAndNoMethod pins the invariant
// soleEmbedded documents, at the level where a constraint is actually decided.
// The interfaces are BUILT rather than parsed: the shapes below include ones
// no Go source spells the same way twice, and go/types is the authority on
// which representation each spelling produces.
func TestSoleEmbeddedNeedsExactlyOneEmbeddingAndNoMethod(t *testing.T) {
	t.Parallel()

	bytes := types.NewSlice(types.Typ[types.Byte])
	sole := types.NewInterfaceType(nil, []types.Type{bytes})
	sole.Complete()
	resolved, ok := soleEmbedded(sole)
	assert.True(t, ok, "one embedded type and no method is what resolves")
	assert.Equal(t, bytes, resolved)

	oneTerm := types.NewUnion([]*types.Term{types.NewTerm(true, bytes)})
	union := types.NewInterfaceType(nil, []types.Type{oneTerm})
	union.Complete()
	resolved, ok = soleEmbedded(union)
	assert.True(t, ok, "a union of exactly one term is the same one type")
	assert.Equal(t, bytes, resolved)

	two := types.NewUnion([]*types.Term{
		types.NewTerm(true, bytes),
		types.NewTerm(true, types.Typ[types.String]),
	})
	wide := types.NewInterfaceType(nil, []types.Type{two})
	wide.Complete()
	_, ok = soleEmbedded(wide)
	assert.False(t, ok, "two terms admit two types, and a caller passes either")

	none := types.NewInterfaceType(nil, nil)
	none.Complete()
	_, ok = soleEmbedded(none)
	assert.False(t, ok, "`any` embeds nothing, so there is nothing to resolve")

	result := types.NewTuple(types.NewParam(token.NoPos, nil, "", types.Typ[types.Int]))
	length := types.NewFunc(token.NoPos, nil, "Len", types.NewSignatureType(nil, nil, nil, nil, result, false))
	withMethod := types.NewInterfaceType([]*types.Func{length}, []types.Type{bytes})
	withMethod.Complete()
	_, ok = soleEmbedded(withMethod)
	assert.False(t, ok, "a method means only DEFINED types satisfy it — the vocabulary the rule exempts")
}
