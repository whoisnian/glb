//go:build aix || linux || solaris || zos

package ansi

import "golang.org/x/sys/unix"

func isSupported(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS)
	return err == nil
}
