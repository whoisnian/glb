package osutil_test

import (
	"os"
	"testing"
	"time"

	"github.com/whoisnian/glb/util/osutil"
)

func TestWaitForInterrupt(t *testing.T) {
	done := make(chan struct{})
	waitingTime := 50 * time.Millisecond
	timeoutTime := 10*time.Millisecond + waitingTime

	go func() {
		osutil.WaitForInterrupt()
		close(done)
	}()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Errorf("os.FindProcess() error: %v", err)
	}
	time.AfterFunc(waitingTime, func() { p.Signal(os.Interrupt) })

	select {
	case <-done:
	case <-time.After(timeoutTime):
		t.Error("WaitForInterrupt timeout")
	}
}
