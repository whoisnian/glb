// Package ioutil implements some I/O utility functions.
package ioutil

import (
	"io"
	"os"
)

// SeekAndReadAll reload file content without reopen the file.
// Usually used for Linux /proc pseudo-filesystem.
func SeekAndReadAll(fi *os.File) ([]byte, error) {
	if _, err := fi.Seek(0, 0); err != nil {
		return []byte{}, err
	}
	return io.ReadAll(fi)
}
