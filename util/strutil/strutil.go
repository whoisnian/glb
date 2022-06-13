// Package strutil implements some string utility functions.
package strutil

import "strings"

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
