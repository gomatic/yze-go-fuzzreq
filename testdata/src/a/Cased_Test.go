package a

// The go tool's test-file rule is CASE-SENSITIVE: `Cased_Test.go` does not end
// in `_test.go`, so `go list` puts it in GoFiles and `go test` never runs a
// Test function declared here. Case-folding the name before matching is the
// ordinary instinct of anyone bitten by a Windows or macOS path, and on this
// case-INSENSITIVE filesystem the file still opens under either spelling — so
// the widening is invisible everywhere except here.

// Cased consumes an untrusted string from a file one shift key away from exempt.
func Cased(raw string) error { // want `consumes untrusted input \(string\)`
	_ = raw
	return nil
}
