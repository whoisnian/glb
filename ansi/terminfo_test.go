package ansi_test

import (
	"os"
	"testing"

	"github.com/whoisnian/glb/ansi"
)

func TestScrollUpN(t *testing.T) {
	var tests = []struct {
		input int
		want  string
	}{
		{-1, ""},
		{0, ""},
		{1, "\x1bM"},
		{2, "\x1bM\x1bM"},
		{5, "\x1bM\x1bM\x1bM\x1bM\x1bM"},
	}
	for _, test := range tests {
		if got := ansi.ScrollUpN(test.input); got != test.want {
			t.Errorf("ScrollUpN(%d) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestScrollDownN(t *testing.T) {
	var tests = []struct {
		input int
		want  string
	}{
		{-1, ""},
		{0, ""},
		{1, "\x1bD"},
		{2, "\x1bD\x1bD"},
		{5, "\x1bD\x1bD\x1bD\x1bD\x1bD"},
	}
	for _, test := range tests {
		if got := ansi.ScrollDownN(test.input); got != test.want {
			t.Errorf("ScrollDownN(%d) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestSetCursorPos(t *testing.T) {
	var tests = []struct {
		input [2]int
		want  string
	}{
		{[2]int{-1, -1}, "\x1b[1;1H"},
		{[2]int{0, 0}, "\x1b[1;1H"},
		{[2]int{-1, 5}, "\x1b[1;5H"},
		{[2]int{5, 5}, "\x1b[5;5H"},
	}
	for _, test := range tests {
		if got := ansi.SetCursorPos(test.input[0], test.input[1]); got != test.want {
			t.Errorf("SetCursorPos(%d, %d) = %q, want %q", test.input[0], test.input[1], got, test.want)
		}
	}
}

func TestSupported(t *testing.T) {
	fi, err := os.CreateTemp("", "TestSupported_")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(fi.Name())
	defer fi.Close()

	if ansi.Supported(int(fi.Fd())) {
		t.Errorf("ansi.Supported(file) = true, want false")
	}
}
