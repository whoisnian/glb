package ansi

import (
	"os"

	"golang.org/x/term"
)

// color
const (
	BlackFG   string = "\x1b[30m"
	RedFG     string = "\x1b[31m"
	GreenFG   string = "\x1b[32m"
	YellowFG  string = "\x1b[33m"
	BlueFG    string = "\x1b[34m"
	MagentaFG string = "\x1b[35m"
	CyanFG    string = "\x1b[36m"
	WhiteFG   string = "\x1b[37m"

	BlackBG   string = "\x1b[40m"
	RedBG     string = "\x1b[41m"
	GreenBG   string = "\x1b[42m"
	YellowBG  string = "\x1b[43m"
	BlueBG    string = "\x1b[44m"
	MagentaBG string = "\x1b[45m"
	CyanBG    string = "\x1b[46m"
	WhiteBG   string = "\x1b[47m"

	ResetAll string = "\x1b[0m"
)

func IsColorSupported() bool {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	} else {
		return false
	}
}
