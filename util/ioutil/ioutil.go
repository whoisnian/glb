// Package ioutil implements some I/O utility functions.
package ioutil

import (
	"io"
	"math/rand/v2"
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

// ReadRand read len(buf) bytes from *rand.Rand and writes them into buf.
// It always returns len(buf) and a nil error.
func ReadRand(r *rand.Rand, buf []byte) (n int, err error) {
	var pos, val uint64
	for n = range buf {
		if pos == 0 {
			val = r.Uint64()
			pos = 8
		}
		buf[n] = byte(val)
		val >>= 8
		pos--
	}
	return n, nil
}
