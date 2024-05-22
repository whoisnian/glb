//go:build !windows && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !aix && !linux && !solaris && !zos

package ansi

// from https://github.com/golang/go/blob/adbfb672ba485630d75f8b5598228a63f4af08a4/src/go/build/syslist.go
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
func isSupported(_ uintptr) bool {
	return false
}
