package fsutil

import (
	"os"
	"testing"
)

func TestResolveHomeDir(t *testing.T) {
	// test for Unix
	env := os.Getenv("HOME")
	defer os.Setenv("HOME", env)
	os.Setenv("HOME", "/home/user")

	var tests = []struct {
		input string
		want  string
	}{
		{"", "."},        // clean only
		{"./doc", "doc"}, // clean only
		{"/tmp", "/tmp"}, // clean only
		{"~/", "/home/user"},
		{"~/.", "/home/user"},
		{"~/..", "/home"},
		{"~/ssh", "/home/user/ssh"},
	}
	for _, test := range tests {
		if got, err := ResolveHomeDir(test.input); err != nil {
			t.Errorf("ResolveHomeDir(%q) error: %v", test.input, err)
		} else if got != test.want {
			t.Errorf("ResolveHomeDir(%q) = %q, want %q", test.input, got, test.want)
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
		if got := ResolveBase(test.inputBase, test.inputRaw); got != test.want {
			t.Errorf("ResolveBase(%q, %q) = %q, want %q", test.inputBase, test.inputRaw, got, test.want)
		}
	}
}
