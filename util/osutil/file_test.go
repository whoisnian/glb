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

	n, err := osutil.CopyFile(srcPath, destPath)
	if !os.IsNotExist(err) || n != 0 {
		t.Fatalf("CopyFile() = (%v, %v), want (%v, %v)", n, err, 0, os.ErrNotExist)
	}

	if err = os.WriteFile(srcPath, want, osutil.DefaultFileMode); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	n, err = osutil.CopyFile(srcPath, destPath)
	if err != nil || int(n) != len(want) {
		t.Fatalf("CopyFile() = (%v, %v), want (%v, %v)", n, err, len(want), nil)
	}

	actual, err := os.ReadFile(destPath)
	if err != nil || !bytes.Equal(actual, want) {
		t.Fatalf("ReadFile() = (%q, %v), want (%q, %v)", actual, err, want, nil)
	}
}

func TestMoveFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "TestMoveFile_")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tempDir)

	want := []byte("hello, world")
	srcPath := filepath.Join(tempDir, "src")
	destPath := filepath.Join(tempDir, "dest")

	if err = osutil.MoveFile(srcPath, destPath); !os.IsNotExist(err) {
		t.Fatalf("MoveFile() = %v, want %v", err, os.ErrNotExist)
	}

	if err = os.WriteFile(srcPath, want, osutil.DefaultFileMode); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err = osutil.MoveFile(srcPath, destPath); err != nil {
		t.Fatalf("MoveFile() = %v, want %v", err, nil)
	}

	actual, err := os.ReadFile(destPath)
	if err != nil || !bytes.Equal(actual, want) {
		t.Fatalf("ReadFile() = (%q, %v), want (%q, %v)", actual, err, want, nil)
	}
}
