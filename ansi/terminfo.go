package ansi

import (
	"fmt"
	"strings"
)

// terminfo
const (
	ClearScreen      string = "\x1b[H\x1b[2J"
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

func ScrollUpN(n int) string {
	return strings.Repeat(ScrollUp, n)
}

func ScrollDownN(n int) string {
	return strings.Repeat(ScrollDown, n)
}

// The position of cursor is 1-based, so (1, 1) means 'top left corner'.
func SetCursorPos(row, col int) string {
	return fmt.Sprintf(SetCursorAddress, row, col)
}
