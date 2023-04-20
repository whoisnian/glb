package osutil

import (
	"io"
	"os"
)

// CopyFile reads data from source file and writes to target file.
// If the target file already exists, it is overwritten.
func CopyFile(srcPath, destPath string) (int64, error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer dest.Close()

	return io.Copy(dest, src)
}
