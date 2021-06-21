package ansi

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
