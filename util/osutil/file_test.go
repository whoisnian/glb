package osutil_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/whoisnian/glb/util/osutil"
)

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "TestCopyFile_")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tempDir)

	want := []byte("hello, world")
	srcPath := filepath.Join(tempDir, "src")
	destPath := filepath.Join(tempDir, "dest")

	err = os.WriteFile(srcPath, want, 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	n, err := osutil.CopyFile(srcPath, destPath)
	if err != nil {
		t.Fatalf("CopyFile: %v", err)
	} else if int(n) != len(want) {
		t.Fatalf("CopyFile: get length %q, want %q", n, len(want))
	}

	actual, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(actual, want) {
		t.Fatalf("CopyFile: get data %q, want %q", actual, want)
	}
}
