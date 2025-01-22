package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"
)

const reTextTime = `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})`

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	// skip if args is empty
	ll := l.With()
	if ll != l {
		t.Fatalf("origin.With() got %v, want %v", ll, l)
	}

	// new Handler
	ll = l.With("a", 1, "b", true)
	ll.Log(context.Background(), LevelInfo, "m")
	got, want := buf.String(), `^time=`+reTextTime+` level=INFO msg=m a=1 b=true\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("new.Handle() got %q, want matched by %s", got, want)
	}

	// new2 Handler
	buf.Reset()
	ll = ll.With(slog.String("c", "str"))
	ll.Log(context.Background(), LevelInfo, "mm")
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=mm a=1 b=true c=str\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("new2.Handle() got %q, want matched by %s", got, want)
	}

	// origin Handler
	buf.Reset()
	l.Log(context.Background(), LevelInfo, "mmm")
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=mmm\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("origin.Handle() got %q, want matched by %s", got, want)
	}
}

func TestLoggerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	// skip if args is empty
	ll := l.WithGroup("")
	if ll != l {
		t.Fatalf("origin.WithGroup(\"\") got %v, want %v", ll, l)
	}

	// new Handler
	ll = l.WithGroup("g")
	ll.Log(context.Background(), LevelInfo, "m", "a", 1)
	got, want := buf.String(), `^time=`+reTextTime+` level=INFO msg=m g.a=1\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("new.Handle() got %q, want matched by %s", got, want)
	}

	// new2 Handler
	buf.Reset()
	ll = ll.WithGroup("c")
	ll.Log(context.Background(), LevelInfo, "mm", "b", false)
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=mm g.c.b=false\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("new2.Handle() got %q, want matched by %s", got, want)
	}

	// origin Handler
	buf.Reset()
	l.Log(context.Background(), LevelInfo, "mmm", "c", "str")
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=mmm c=str\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("origin.Handle() got %q, want matched by %s", got, want)
	}
}

func TestLoggerRace(t *testing.T) {
	const P = 10
	const N = 10000
	done := make(chan struct{})
	l := New(NewTextHandler(io.Discard, Options{LevelInfo, true, true}))
	for i := 0; i < P; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < N; j++ {
				l.Log(context.Background(), LevelInfo, "mm", "a", 1)
				ll := l.With(slog.Bool("b", false))
				ll.Log(context.Background(), LevelInfo, "mm", "a", 1)
				ll = l.WithGroup("g")
				ll = ll.With(slog.String("c", "str"))
				ll.Log(context.Background(), LevelInfo, "mm", "a", 1)
			}
		}()
	}
	for i := 0; i < P; i++ {
		<-done
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	// Info
	l.Info(context.Background(), "msg", "a", 1, "b", 2)
	got, want := buf.String(), `^time=`+reTextTime+` level=INFO msg=msg a=1 b=2\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Info() got %q, want matched by %s", got, want)
	}

	// Debug
	buf.Reset()
	l.Debug(context.Background(), "ddd", "a", 1)
	if buf.String() != "" {
		t.Errorf("Logger.Debug() got %q, want %q", got, "")
	}

	// Warn
	buf.Reset()
	l.Warn(context.Background(), "w", slog.Duration("dur", 3*time.Second))
	got, want = buf.String(), `^time=`+reTextTime+` level=WARN msg=w dur=3s\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Warn() got %q, want matched by %s", got, want)
	}

	// Error
	buf.Reset()
	l.Error(context.Background(), "bad", slog.Int("a", 1), "missing")
	got, want = buf.String(), `^time=`+reTextTime+` level=ERROR msg=bad a=1 !BADKEY=missing\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Error() got %q, want matched by %s", got, want)
	}

	// Log
	buf.Reset()
	l.Log(context.Background(), LevelInfo, "log", "a", 1)
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=log a=1\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Log() got %q, want matched by %s", got, want)
	}

	// LogAttrs
	buf.Reset()
	l.LogAttrs(context.Background(), LevelInfo, "logattrs", slog.Bool("b", true), slog.Float64("c", 0.25))
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=logattrs b=true c=0.25\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.LogAttrs() got %q, want matched by %s", got, want)
	}
}

func TestLoggerOutputf(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	// Infof
	l.Infof(context.Background(), "from:%s", "192.168.0.2")
	got, want := buf.String(), `^time=`+reTextTime+` level=INFO msg=from:192.168.0.2\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Infof() got %q, want matched by %s", got, want)
	}

	// Debugf
	buf.Reset()
	l.Debugf(context.Background(), "cost:%ds", 5)
	if buf.String() != "" {
		t.Errorf("Logger.Debugf() got %q, want %q", got, "")
	}

	// Warnf
	buf.Reset()
	l.Warnf(context.Background(), "cost:%.2fs", 2.5)
	got, want = buf.String(), `^time=`+reTextTime+` level=WARN msg=cost:2.50s\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Warnf() got %q, want matched by %s", got, want)
	}

	// Errorf
	buf.Reset()
	l.Errorf(context.Background(), "err:%s", errors.New("invalid"))
	got, want = buf.String(), `^time=`+reTextTime+` level=ERROR msg=err:invalid\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Errorf() got %q, want matched by %s", got, want)
	}

	// Logf
	buf.Reset()
	l.Logf(context.Background(), LevelInfo, "finished:%d%%", 80)
	got, want = buf.String(), `^time=`+reTextTime+` level=INFO msg=finished:80%\n$`
	if !regexp.MustCompile(want).MatchString(got) {
		t.Errorf("Logger.Logf() got %q, want matched by %s", got, want)
	}
}

func TestLoggerFatal(t *testing.T) {
	// https://stackoverflow.com/a/33404435/11239247
	if os.Getenv("TEST_FATAL") == "true" {
		l := New(NewTextHandler(os.Stderr, Options{LevelInfo, false, false}))
		l.Fatal(context.Background(), "f", "t", testTime)
		return
	}
	want := `^time=` + reTextTime + ` level=FATAL msg=f t=2000-01-02T03:04:05Z\n$`
	var stderr bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestLoggerFatal")
	cmd.Env = append(os.Environ(), "TEST_FATAL=true")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
		got := stderr.String()
		if !regexp.MustCompile(want).MatchString(got) {
			t.Fatalf("Logger.Fatal() got %q, want matched by %s", got, want)
		}
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestLoggerFatalf(t *testing.T) {
	// https://stackoverflow.com/a/33404435/11239247
	if os.Getenv("TEST_FATAL") == "true" {
		l := New(NewTextHandler(os.Stderr, Options{LevelInfo, false, false}))
		l.Fatalf(context.Background(), "%[3]*.[2]*[1]f", 12.0, 2, 6)
		return
	}
	want := `^time=` + reTextTime + ` level=FATAL msg=" 12.00"\n$`
	var stderr bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestLoggerFatalf")
	cmd.Env = append(os.Environ(), "TEST_FATAL=true")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
		got := stderr.String()
		if !regexp.MustCompile(want).MatchString(got) {
			t.Fatalf("Logger.Fatalf() got %q, want matched by %s", got, want)
		}
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
