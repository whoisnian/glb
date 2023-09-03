// Package ansi defines frequently-used ANSI escape sequences.
package ansi

// IsSupported reports if the ANSI escape sequences are supported. Example:
//
//	if ansi.IsSupported(os.Stdout.Fd()) {
//		fmt.Println(ansi.RedFG, "ERROR", ansi.Reset)
//	}
func IsSupported(fd uintptr) bool {
	return isSupported(fd)
}
