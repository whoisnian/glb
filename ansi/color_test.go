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
	{"BlackFG", ansi.BlackFG, escape.Black},
	{"RedFG", ansi.RedFG, escape.Red},
	{"GreenFG", ansi.GreenFG, escape.Green},
	{"YellowFG", ansi.YellowFG, escape.Yellow},
	{"BlueFG", ansi.BlueFG, escape.Blue},
	{"MagentaFG", ansi.MagentaFG, escape.Magenta},
	{"CyanFG", ansi.CyanFG, escape.Cyan},
	{"WhiteFG", ansi.WhiteFG, escape.White},
	{"Reset", ansi.Reset, escape.Reset},
}

func TestColorFG(t *testing.T) {
	for _, test := range colorTests {
		if test.ansi != string(test.term) {
			t.Errorf("ansi.%s: got %q, want %q", test.name, test.ansi, test.term)
		}
	}
}
