package logger_test

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/whoisnian/glb/logger"
)

const (
	Rdate       = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime       = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	RplainLabel = `\[(I|W|E|D|R)\]`
	RcolorLabel = `\x1b\[3[0-9]m\[(I|W|E|D|R)\]\x1b\[0m`
	Text        = `hello world`
)

var (
	reForPlainLabel = regexp.MustCompile("^" + Rdate + " " + Rtime + " " + RplainLabel + " " + Text + `\n$`)
	reForColorLabel = regexp.MustCompile("^" + Rdate + " " + Rtime + " " + RcolorLabel + " " + Text + `\n$`)
)

func resetLogger() {
	logger.SetOutput(os.Stderr)
	logger.SetDebug(false)
	logger.SetColorful(false)
}

func TestOutput(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	buf.Reset()
	logger.Info(Text)
	if !reForPlainLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}

	buf.Reset()
	logger.Warn(Text)
	if !reForPlainLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}

	buf.Reset()
	logger.Error(Text)
	if !reForPlainLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	buf.Reset()
	logger.Debug(Text)
	if logger.IsDebug() || buf.Len() != 0 {
		t.Fatal("debug log should be disable if debug=false")
	}

	buf.Reset()
	logger.SetDebug(true)
	logger.Debug(Text)
	if !reForPlainLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}
}

func TestColorful(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	buf.Reset()
	logger.Info(Text)
	if logger.IsColorful() || !reForPlainLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}

	buf.Reset()
	logger.SetColorful(true)
	logger.Info(Text)
	if !reForColorLabel.Match(buf.Bytes()) {
		t.Fatalf("log should match %q is %q", reForPlainLabel, buf.Bytes())
	}
}

func TestPanic(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	defer func() {
		if recover() == nil || !reForPlainLabel.Match(buf.Bytes()) {
			t.Fatalf("panic log should match %q is %q", reForPlainLabel, buf.Bytes())
		}
	}()

	logger.Panic(Text)
	t.Fatal("logger Panic() should get panic")
}

func TestFatal(t *testing.T) {
	// https://stackoverflow.com/a/33404435/11239247
	if os.Getenv("TEST_FATAL") == "true" {
		logger.Fatal(Text)
		return
	}
	var stderr bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "TEST_FATAL=true")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
		if !reForPlainLabel.Match(stderr.Bytes()) {
			t.Fatalf("fatal log should match %q is %q", reForPlainLabel, stderr.Bytes())
		}
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
