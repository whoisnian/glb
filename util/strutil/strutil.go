// Package strutil implements some string utility functions.
package strutil

// SliceContain determines whether a string slice includes a certain value.
func SliceContain(slice []string, value string) bool {
	for i := range slice {
		if value == slice[i] {
			return true
		}
	}
	return false
}
