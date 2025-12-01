// nolint
package rarchive

import (
	"compress/gzip"
	"os"
)

// IsTarGzipFile checks if the file at the given path is a gzipped tarball
// by attempting to open it as such.
func IsTarGzipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Attempt to create a gzip reader
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer gzr.Close()

	return true
}
