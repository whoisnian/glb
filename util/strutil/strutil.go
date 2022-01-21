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
