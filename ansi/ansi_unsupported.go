//go:build !windows && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !aix && !linux && !solaris && !zos

package ansi

// from https://github.com/golang/go/blob/e7fbd28a4dbf92721f040dfb2c877153333054d1/src/go/build/syslist.go
// ansi_windows.go
// * windows
// ansi_unix_bsd.go
// * darwin
// * dragonfly
// * freebsd
// * netbsd
// * openbsd
// ansi_unix_other.go
// * aix
// * linux
// * solaris
// * zos
// ansi_unsupported.go
// * android
// * hurd
// * illumos
// * ios
// * js
// * nacl
// * plan9
// * wasip1
func isSupported(fd uintptr) bool {
	return false
}
