//go:build windows

package ansi

import "golang.org/x/sys/windows"

// from https://github.com/golang/term/blob/46c790f81f1f50148a57f7ddf0c637b84ff2f0e6/term_windows.go#L17
// from https://github.com/fatih/color/blob/b6598b12a645b3159c1bce51b0e3fafc269510be/color_windows.go#L9
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
