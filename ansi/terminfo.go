package ansi

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// terminfo
const (
	ClearScreen      string = "\x1b[H\x1b[3J"
	ClearLineToRight string = "\x1b[K"
	ScrollUp         string = "\x1bM"
	ScrollDown       string = "\x1bD"
	GetCursorAddress string = "\x1b[6n"
	SetCursorAddress string = "\x1b[%d;%dH"
	CursorVisible    string = "\x1b[?12;25h"
	CursorInvisible  string = "\x1b[?25l"
	EnableLineWrap   string = "\x1b[?7h"
	DisableLineWrap  string = "\x1b[?7l"
)

// ScrollUpN repeats ScrollUp for N times.
func ScrollUpN(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(ScrollUp, n)
}

// ScrollDownN repeats ScrollDown for N times.
func ScrollDownN(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(ScrollDown, n)
}

// SetCursorPos moves cursor to (row, col) location.
// The position of cursor is 1-based, so (1, 1) means 'top left corner'.
func SetCursorPos(row, col int) string {
	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}
	return fmt.Sprintf(SetCursorAddress, row, col)
}

// Supported reports if the ANSI escape sequences are supported. Example:
//
//	if ansi.Supported(int(os.Stdout.Fd())) {
//		fmt.Println(ansi.RedFG, "ERROR", ansi.Reset)
//	}
func Supported(fd int) bool {
	if term.IsTerminal(fd) {
		if runtime.GOOS == "windows" {
			// https://github.com/microsoft/terminal/issues/1040
			_, ok := os.LookupEnv("WT_SESSION")
			return ok
		}
		return true
	} else {
		return false
	}
}
