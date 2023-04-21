// Package strutil implements some string utility functions.
package strutil

import (
	"strings"
	"unsafe"
)

// SliceContain determines whether a string slice includes a certain value.
func SliceContain(slice []string, value string) bool {
	for i := range slice {
		if value == slice[i] {
			return true
		}
	}
	return false
}

// ShellEscape escapes a string for use in a shell command.
func ShellEscape(s string) string {
	return "'" + strings.Replace(s, "'", `'"'"'`, -1) + "'"
}

// ShellEscapeExceptTilde escapes a string for use in a shell command, except '~'.
func ShellEscapeExceptTilde(s string) string {
	if strings.HasPrefix(s, "~/") {
		return "~/" + ShellEscape(s[2:])
	}
	return ShellEscape(s)
}

// IsDigitString checks if a string only contains digits.
func IsDigitString(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
}

// UnsafeStringToBytes converts string to byte slice without allocation.
func UnsafeStringToBytes(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeBytesToString converts byte slice to string without allocation.
// The bytes passed to function must not be modified afterwards.
func UnsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
