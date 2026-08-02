package fuzzreq

import (
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/assert"
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
