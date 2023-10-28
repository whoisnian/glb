package fsutil_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/whoisnian/glb/util/fsutil"
)

func TestResolveHomeDir(t *testing.T) {
	homeKey, homeVal := "HOME", "/home/nian"
	if runtime.GOOS == "windows" {
		homeKey = "USERPROFILE"
		homeVal = `C:\Users\nian`
	}
	env := os.Getenv(homeKey)
	defer os.Setenv(homeKey, env)
	os.Setenv(homeKey, homeVal)

	var tests = []struct {
		input string
		want  string
	}{
		{"", "."},        // clean only
		{"./doc", "doc"}, // clean only
		{"/tmp", "/tmp"}, // clean only
		{"~/", homeVal},
		{"~/.", homeVal},
		{"~/..", filepath.Dir(homeVal)},
		{"~/ssh", homeVal + "/ssh"},
	}
	for _, test := range tests {
		want := test.want
		if runtime.GOOS == "windows" {
			want = filepath.FromSlash(test.want)
		}
		if got, err := fsutil.ResolveHomeDir(test.input); err != nil {
			t.Errorf("ResolveHomeDir(%q) error: %v", test.input, err)
		} else if got != want {
			t.Errorf("ResolveHomeDir(%q) = %q, want %q", test.input, got, want)
		}
	}
}

func TestResolveBase(t *testing.T) {
	var tests = []struct {
		inputBase string
		inputRaw  string
		want      string
	}{
		{"/data", "", "/data"},
		{"/data", ".", "/data"},
		{"/data", "..", "/data"},
		{"/data", "doc", "/data/doc"},
		{"/data", "../doc", "/data/doc"},
		{"/data", "/doc", "/data/doc"},
		{"/data", "/doc/..", "/data"},
	}
	for _, test := range tests {
		want := test.want
		if runtime.GOOS == "windows" {
			want = filepath.FromSlash(test.want)
		}
		if got := fsutil.ResolveBase(test.inputBase, test.inputRaw); got != want {
			t.Errorf("ResolveBase(%q, %q) = %q, want %q", test.inputBase, test.inputRaw, got, want)
		}
	}
}
