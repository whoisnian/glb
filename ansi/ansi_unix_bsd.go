//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package ansi

import "golang.org/x/sys/unix"

func isSupported(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TIOCGETA)
	return err == nil
}
