// Package osutil implements some OS utility functions.
package osutil

import (
	"os"
	"os/signal"
)

// WaitForInterrupt blocks until current process receives SIGINT.
func WaitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	<-c
}
