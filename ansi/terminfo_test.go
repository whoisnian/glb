package ansi

import (
	"testing"
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
		if got := ScrollUpN(test.input); got != test.want {
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
		if got := ScrollDownN(test.input); got != test.want {
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
		if got := SetCursorPos(test.input[0], test.input[1]); got != test.want {
			t.Errorf("SetCursorPos(%d, %d) = %q, want %q", test.input[0], test.input[1], got, test.want)
		}
	}
}
