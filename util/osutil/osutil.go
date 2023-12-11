// Package osutil implements some OS utility functions.
package osutil

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitFor blocks until current process receives any of signals.
func WaitFor(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	defer signal.Stop(c)

	<-c
}

// WaitForInterrupt blocks until current process receives SIGINT.
func WaitForInterrupt() {
	WaitFor(syscall.SIGINT)
}

// WaitForStop blocks until current process receives SIGINT or SIGTERM.
func WaitForStop() {
	WaitFor(syscall.SIGINT, syscall.SIGTERM)
}
