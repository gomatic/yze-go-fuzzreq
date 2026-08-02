package fuzzreq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	fuzzreq "github.com/gomatic/yze-go-fuzzreq"
)

// TestUnfuzzedEntryPoints pins both directions: the bare []byte/string/
// io.Reader entry points in the fuzz-free package report; the named-type,
// error-free, and unexported shapes do not; and the package WITH a fuzz
// target owes nothing.
func TestUnfuzzedEntryPoints(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), fuzzreq.Analyzer, "a", "b")
}

// TestRegistrationIsWellFormed pins the yze wiring.
func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, fuzzreq.Registration.Validate())
	assert.Equal(t, "yze/fuzzreq", fuzzreq.Registration.RuleID())
	assert.Same(t, fuzzreq.Analyzer, fuzzreq.Registration.Analyzer)
}
