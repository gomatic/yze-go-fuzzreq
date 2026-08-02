// Package b holds the same surface WITH a fuzz target: nothing is owed.
package b

// Parse consumes bare bytes and can fail.
func Parse(raw []byte) (string, error) {
	return string(raw), nil
}
