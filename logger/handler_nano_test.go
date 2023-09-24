package logger

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

var testTime = time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)

func TestNanoHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewNanoHandler(&buf, NewOptions(LevelInfo, false, false))

	// skip if attrs is empty
	hh := h.WithAttrs([]slog.Attr{})
	if hh != h {
		t.Fatalf("origin.WithAttrs([]) got %v, want %v", hh, h)
	}

	hh = h.WithAttrs([]slog.Attr{slog.Duration("dur", time.Minute), slog.Bool("b", true)})
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Int("a", 1))

	// new Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new.Handle() error: %v", err)
	}
	got, want := buf.String(), "2000-01-02 03:04:05 [I] m 1m0s true 1\n"
	if got != want {
		t.Errorf("new.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), "2000-01-02 03:04:05 [I] m 1\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestNanoHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewNanoHandler(&buf, NewOptions(LevelInfo, false, false))

	hh := h.WithGroup("s")
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Duration("dur", time.Millisecond))
	r.AddAttrs(slog.Int("a", 2))

	// new Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new.Handle() error: %v", err)
	}
	got, want := buf.String(), "2000-01-02 03:04:05 [I] m 1ms 2\n"
	if got != want {
		t.Errorf("new.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), "2000-01-02 03:04:05 [I] m 1ms 2\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestNanoHandler(t *testing.T) {
}

func TestNanoHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := 0; i < 10; i++ {
		r.AddAttrs(slog.Int("x = y", i))
	}
	var h Handler = NewNanoHandler(io.Discard, NewOptions(LevelInfo, false, false))
	got := int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("origin.Handle() got %d allocs, want 0", got)
	}

	h = h.WithGroup("s")
	r.AddAttrs(slog.Group("g", slog.Int("a", 1)))
	got = int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("new.Handle() got %d allocs, want 0", got)
	}
}

func TestAppendNanoValue(t *testing.T) {
	var tests = []struct {
		input slog.Value
		want  string
	}{
		{slog.StringValue("abc def"), " abc def"},
		{slog.IntValue(123), " 123"},
		{slog.Int64Value(-456), " -456"},
		{slog.Uint64Value(789), " 789"},
		{slog.Float64Value(0.125), " 0.125"},
		{slog.BoolValue(false), " false"},
		{slog.BoolValue(true), " true"},
		{slog.TimeValue(testTime), " 2000-01-02T03:04:05Z"},
		{slog.DurationValue(time.Minute), " 1m0s"},
		{slog.DurationValue(time.Microsecond), " 1µs"},
		{slog.GroupValue(
			slog.String("method", "GET"),
			slog.Int("code", 200),
		), " GET 200"},
		{slog.AnyValue([]byte("test")), " [116 101 115 116]"},
		{slog.AnyValue(map[string]int{"age": 18}), " map[age:18]"},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		buf = buf[:0]
		appendNanoValue(&buf, test.input)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendNanoValue(%+v) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestAppendNanoSource(t *testing.T) {
	var tests = []struct {
		depth  int
		prefix string
	}{
		{0, "runtime/extern.go:"},
		{1, "logger/handler_nano_test.go:"},
		{2, "testing/testing.go:"},
	}
	buf := make([]byte, 64)
	for _, test := range tests {
		var (
			pc  uintptr
			pcs [1]uintptr
		)
		runtime.Callers(test.depth, pcs[:])
		pc = pcs[0]

		buf = buf[:0]
		appendNanoSource(&buf, pc)

		got := string(buf)
		if !strings.HasPrefix(got, test.prefix) || !regexp.MustCompile(`\d+`).MatchString(got[len(test.prefix):]) {
			t.Fatalf("appendNanoSource(%d) = %s, want %sxxx", test.depth, got, test.prefix)
		}
	}
}

func TestAppendDateTime(t *testing.T) {
	buf := make([]byte, 32)
	for _, tm := range []time.Time{
		time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		time.Date(2000, 1, 2, 3, 4, 5, 400, time.Local),
		time.Date(2000, 11, 12, 3, 4, 500, 5e7, time.UTC),
	} {
		buf = buf[:0]
		appendDateTime(&buf, tm)

		got, want := string(buf), tm.Format(time.DateTime)
		if got != want {
			t.Fatalf("appendDateTime(%v) = %s, want %s", tm, got, want)
		}
	}
}
