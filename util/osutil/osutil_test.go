package osutil

import (
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestWaitForInterrupt(t *testing.T) {
	done := make(chan struct{})
	waitingTime := 50 * time.Millisecond
	timeoutTime := 10*time.Millisecond + waitingTime

	go func() {
		WaitForInterrupt()
		close(done)
	}()

	time.AfterFunc(waitingTime, func() {
		// test for Unix
		unix.Kill(unix.Getpid(), unix.SIGINT)
	})

	select {
	case <-done:
	case <-time.After(timeoutTime):
		t.Error("WaitForInterrupt timeout")
	}
}
