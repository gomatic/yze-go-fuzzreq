//line notatest.go:1
package a

// The other direction of the same directive: a file the go tool compiles ONLY
// into the test binary, claiming an ordinary source name. Reading the claimed
// name would report a declaration no importer can reach and no fuzzer can
// drive, which is a finding its author cannot act on. Nothing is reported here.

// Disguise is the export_test idiom wearing a production filename.
func Disguise(raw string) error {
	_ = raw
	return nil
}
