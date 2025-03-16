package osutil_test

import (
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/whoisnian/glb/util/osutil"
)

const (
	waitingTime = 50 * time.Millisecond
	timeoutTime = 100 * time.Millisecond
)

func TestWaitForInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("sending syscall.SIGINT is not implemented for windows")
	}

	done := make(chan struct{})
	go func() {
		osutil.WaitForInterrupt()
		close(done)
	}()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("os.FindProcess() error: %v", err)
	}

	time.AfterFunc(waitingTime, func() { p.Signal(syscall.SIGINT) })
	select {
	case <-done:
	case <-time.After(timeoutTime):
		t.Fatal("WaitForInterrupt timeout")
	}
}

func TestWaitForStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("sending syscall.SIGTERM is not implemented for windows")
	}

	done := make(chan struct{})
	go func() {
		osutil.WaitForStop()
		close(done)
	}()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("os.FindProcess() error: %v", err)
	}

	time.AfterFunc(waitingTime, func() { p.Signal(syscall.SIGTERM) })
	select {
	case <-done:
	case <-time.After(timeoutTime):
		t.Fatal("WaitForStop timeout")
	}
}
