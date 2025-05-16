// Package strutil implements some string utility functions.
package strutil

import (
	"strings"
	"unsafe"
)

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

// IsDigitString checks if a string only contains ASCII digits.
// It returns false for empty string.
func IsDigitString(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// Camelize converts an ASCII string to camel case like string, and supports specifying the case of the first letter.
func Camelize(s string, upper bool) string {
	var buf []byte
	for i := range len(s) {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z':
			if upper {
				c -= 'a' - 'A' // lowercase to uppercase
				upper = false
			}
			buf = append(buf, c)
		case 'A' <= c && c <= 'Z':
			if !upper {
				c += 'a' - 'A' // uppercase to lowercase
			}
			upper = false
			buf = append(buf, c)
		case '0' <= c && c <= '9':
			buf = append(buf, c)
		default:
			upper = len(buf) > 0 || upper // keep `upper` for first letter in buf
		}
	}
	return string(buf)
}

// Underscore converts an ASCII string to snake case like string, and supports specifying the case of all letters.
func Underscore(s string, upper bool) string {
	const (
		initial     = iota
		upperLetter // [A-Z]
		lowerLetter // [a-z]
		notAlphanum // [^A-Za-z0-9]
	)
	var last int
	var buf []byte
	for i := range len(s) {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z':
			if upper {
				c -= 'a' - 'A' // lowercase to uppercase
			}
			if len(buf) > 0 && last == notAlphanum {
				buf = append(buf, '_', c)
			} else {
				buf = append(buf, c)
			}
			last = lowerLetter
		case 'A' <= c && c <= 'Z':
			if !upper {
				c += 'a' - 'A' // uppercase to lowercase
			}
			if len(buf) > 0 && (last == lowerLetter || last == notAlphanum) {
				buf = append(buf, '_', c)
			} else if len(buf) > 0 && i+1 < len(s) && 'a' <= s[i+1] && s[i+1] <= 'z' { // next letter is lowercase
				buf = append(buf, '_', c) // ABc => A_Bc
			} else {
				buf = append(buf, c)
			}
			last = upperLetter
		case '0' <= c && c <= '9':
			if len(buf) > 0 && last == notAlphanum {
				buf = append(buf, '_', c)
			} else {
				buf = append(buf, c)
			}
			last = initial
		default:
			last = notAlphanum
		}
	}
	return string(buf)
}

// UnsafeStringToBytes converts string to byte slice without allocation.
func UnsafeStringToBytes(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}
	// like https://cs.opensource.google/go/go/+/refs/tags/go1.21.1:src/os/file.go;l=253
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeBytesToString converts byte slice to string without allocation.
// The bytes passed to function must not be modified afterwards.
func UnsafeBytesToString(b []byte) string {
	// like https://cs.opensource.google/go/go/+/refs/tags/go1.21.1:src/strings/builder.go;l=48
	return unsafe.String(unsafe.SliceData(b), len(b))
}
