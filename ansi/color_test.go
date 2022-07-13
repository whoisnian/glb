package ansi_test

import (
	"testing"

	"github.com/whoisnian/glb/ansi"
	"golang.org/x/term"
)

var escape = term.NewTerminal(nil, "").Escape
var colorTests = []struct {
	name string
	ansi string
	term []byte
}{
	{"Black", ansi.Black, escape.Black},
	{"Red", ansi.Red, escape.Red},
	{"Green", ansi.Green, escape.Green},
	{"Yellow", ansi.Yellow, escape.Yellow},
	{"Blue", ansi.Blue, escape.Blue},
	{"Magenta", ansi.Magenta, escape.Magenta},
	{"Cyan", ansi.Cyan, escape.Cyan},
	{"White", ansi.White, escape.White},
	{"Reset", ansi.Reset, escape.Reset},
}

func TestColor(t *testing.T) {
	for _, test := range colorTests {
		if test.ansi != string(test.term) {
			t.Errorf("ansi.%s: got %q, want %q", test.name, test.ansi, test.term)
		}
	}
}
