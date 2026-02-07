package fsutil_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/whoisnian/glb/util/fsutil"
)

func TestExpandHomeDir(t *testing.T) {
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
		{"~tmp", "~tmp"}, // clean only
		{"~", homeVal},
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
		if got, err := fsutil.ExpandHomeDir(test.input); err != nil {
			t.Errorf("ExpandHomeDir(%q) error: %v", test.input, err)
		} else if got != want {
			t.Errorf("ExpandHomeDir(%q) = %q, want %q", test.input, got, want)
		}
	}
}

func TestResolveUrlPath(t *testing.T) {
	var tests = []struct {
		inputBase string
		inputRaw  string
		want      string
	}{
		{"/var/www/html", "", "/var/www/html"},
		{"/var/www/html", ".", "/var/www/html"},
		{"/var/www/html", "..", "/var/www/html"},
		{"/var/www/html", "/", "/var/www/html"},

		{"/var/www/html", "static/image.png", "/var/www/html/static/image.png"},
		{"/var/www/html", "/static/image.png", "/var/www/html/static/image.png"},
		{"/var/www/html", "///static/image.png", "/var/www/html/static/image.png"},
		{"/var/www/html", "/static/./image.png", "/var/www/html/static/image.png"},

		{"/var/www/html", "../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "../../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "../../../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/../../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/../../../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/static/../etc/passwd", "/var/www/html/etc/passwd"},

		{"/var/www/html", "/static\\..\\..\\etc\\passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "\\static\\..\\..\\etc\\passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/static/%2e%2e/file.txt", "/var/www/html/static/%2e%2e/file.txt"},
		{"/var/www/html", "/static\x00/../../etc/passwd", "/var/www/html/etc/passwd"},
		{"/var/www/html", "/static/../..\x00/etc/passwd", "/var/www/html/..\x00/etc/passwd"},
	}
	for _, test := range tests {
		want := test.want
		if runtime.GOOS == "windows" {
			want = filepath.FromSlash(test.want)
		}
		if got := fsutil.ResolveUrlPath(test.inputBase, test.inputRaw); got != want {
			t.Errorf("ResolveUrlPath(%q, %q) = %q, want %q", test.inputBase, test.inputRaw, got, want)
		}
	}
}
