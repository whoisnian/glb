//go:build windows

package ansi

import "golang.org/x/sys/windows"

// from https://github.com/golang/term/blob/f413282cd8dbb55102093d9f16ab3ba90f7b9b31/term_windows.go#L17
// from https://github.com/fatih/color/blob/d5c210ca2a0ed2ce7d8e46320ef777d64f38c83a/color_windows.go#L9
// from https://learn.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences
func isSupported(fd uintptr) bool {
	var mode, flag uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &mode); err != nil {
		return false
	}

	flag = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.ENABLE_PROCESSED_OUTPUT
	if mode&flag == flag {
		return true
	}

	return windows.SetConsoleMode(windows.Handle(fd), mode|flag) == nil
}
