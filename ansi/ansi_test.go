package ansi_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/whoisnian/glb/ansi"
)

func TestIsSupportedFile(t *testing.T) {
	fi, err := os.CreateTemp("", "TestIsSupportedFile_")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(fi.Name())
	defer fi.Close()

	if ansi.IsSupported(fi.Fd()) {
		t.Errorf("ansi.IsSupported(%s) = true, want false", fi.Name())
	}
}

func TestIsSupportedTerm(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
	}

	fi, err := os.Open("/dev/ptmx")
	if err != nil {
		t.Fatalf("Open(/dev/ptmx): %v", err)
	}
	defer fi.Close()

	if !ansi.IsSupported(fi.Fd()) {
		t.Errorf("ansi.IsSupported(%s) = false, want true", fi.Name())
	}
}
