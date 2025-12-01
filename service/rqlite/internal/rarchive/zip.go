// nolint
package rarchive

import (
	"bytes"
	"os"
)

// IsZipFile checks if the file at the given path is a ZIP file
// by verifying the magic number (the first few bytes of the file).
func IsZipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read the first 4 bytes of the file to check the magic number
	buf := make([]byte, 4)
	_, err = f.Read(buf)
	if err != nil {
		return false
	}

	// ZIP files start with "PK\x03\x04"
	return bytes.Equal(buf, []byte("PK\x03\x04"))
}
