package ansi

import (
	"testing"

	"golang.org/x/term"
)

var escape = term.NewTerminal(nil, "").Escape
var colorTests = []struct {
	name string
	ansi string
	term []byte
}{
	{"Black", Black, escape.Black},
	{"Red", Red, escape.Red},
	{"Green", Green, escape.Green},
	{"Yellow", Yellow, escape.Yellow},
	{"Blue", Blue, escape.Blue},
	{"Magenta", Magenta, escape.Magenta},
	{"Cyan", Cyan, escape.Cyan},
	{"White", White, escape.White},
	{"Reset", Reset, escape.Reset},
}

func TestColor(t *testing.T) {
	for _, test := range colorTests {
		if test.ansi != string(test.term) {
			t.Errorf("ansi.%s: got %q, want %q", test.name, test.ansi, test.term)
		}
	}
}
