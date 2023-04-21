package strutil_test

import (
	"bytes"
	"testing"

	"github.com/whoisnian/glb/util/strutil"
)

func TestSliceContain(t *testing.T) {
	var tests = []struct {
		inputSlice []string
		inputValue string
		want       bool
	}{
		{nil, "", false},
		{[]string{}, "", false},
		{[]string{""}, "", true},
		{[]string{"a"}, "a", true},
		{[]string{"a", "b", "c"}, "c", true},
		{[]string{"a", "b", "c"}, "d", false},
	}
	for _, test := range tests {
		if got := strutil.SliceContain(test.inputSlice, test.inputValue); got != test.want {
			t.Errorf("SliceContain(%q, %q) = %v, want %v", test.inputSlice, test.inputValue, got, test.want)
		}
	}
}

func TestShellEscape(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{``, `''`},
		{`/tmp/dir 1`, `'/tmp/dir 1'`},
		{`"/tmp/dir 1"`, `'"/tmp/dir 1"'`},
		{`'/tmp/dir 1'`, `''"'"'/tmp/dir 1'"'"''`},
		{`~/doc`, `'~/doc'`},
	}
	for _, test := range tests {
		if got := strutil.ShellEscape(test.input); got != test.want {
			t.Errorf("ShellEscape(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestShellEscapeExceptTilde(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{``, `''`},
		{`/tmp/dir 1`, `'/tmp/dir 1'`},
		{`"/tmp/dir 1"`, `'"/tmp/dir 1"'`},
		{`'/tmp/dir 1'`, `''"'"'/tmp/dir 1'"'"''`},
		{`~/doc`, `~/'doc'`},
	}
	for _, test := range tests {
		if got := strutil.ShellEscapeExceptTilde(test.input); got != test.want {
			t.Errorf("ShellEscapeExceptTilde(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestIsDigitString(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"", false},
		{"-1", false},
		{"0.1", false},
		{"1e5", false},
		{"0", true},
		{"001", true},
		{"12345", true},
	}
	for _, test := range tests {
		if got := strutil.IsDigitString(test.input); got != test.want {
			t.Errorf("IsDigitString(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

func TestUnsafeStringToBytes(t *testing.T) {
	var inputs = []string{
		"",
		" ",
		"hello, world",
		"\a\b\\\\t\n\r\"'",
		"\x00\x01\x02\x03\x04",
	}
	for _, input := range inputs {
		if got := strutil.UnsafeStringToBytes(input); !bytes.Equal(got, []byte(input)) {
			t.Errorf("UnsafeStringToBytes(%q) = %v, want %v", input, got, []byte(input))
		}
	}
}

func TestUnsafeBytesToString(t *testing.T) {
	var inputs = [][]byte{
		nil,
		{},
		{'h', 'e', 'l', 'l', 'o'},
		{0, 1, 2, 3, 4},
	}
	for _, input := range inputs {
		if got := strutil.UnsafeBytesToString(input); got != string(input) {
			t.Errorf("UnsafeBytesToString(%q) = %q, want %q", input, got, string(input))
		}
	}
}
