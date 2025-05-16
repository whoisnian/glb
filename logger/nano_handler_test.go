package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/whoisnian/glb/ansi"
)

var testTime = time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)

func TestNanoHandlerEnabled(t *testing.T) {
	for _, test := range levelTests {
		got := NewNanoHandler(io.Discard, Options{Level: test.opts}).Enabled(context.Background(), test.input)
		if got != test.want {
			t.Errorf("(%v).Enabled(ctx, %d) = %v, want %v", test.opts, test.input, got, test.want)
		}
	}
}

func TestNanoHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h slog.Handler = NewNanoHandler(&buf, Options{LevelInfo, false, false})

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
	var h slog.Handler = NewNanoHandler(&buf, Options{LevelInfo, false, false})

	hh := h.WithGroup("s")
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Duration("dur", time.Millisecond))
	r.AddAttrs(slog.Int("a", 2))

	// new1 Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new1.Handle() error: %v", err)
	}
	got, want := buf.String(), "2000-01-02 03:04:05 [I] m 1ms 2\n"
	if got != want {
		t.Errorf("new1.Handle() got %q, want %q", got, want)
	}

	// new2 Handler
	hh = hh.WithAttrs([]slog.Attr{slog.Bool("b", true)})
	hh = hh.WithAttrs([]slog.Attr{slog.Int("c", 3)})
	buf.Reset()
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new2.Handle() error: %v", err)
	}
	got, want = buf.String(), `2000-01-02 03:04:05 [I] m true 3 1ms 2`+"\n"
	if got != want {
		t.Errorf("new2.Handle() got %q, want %q", got, want)
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
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])
	f, _ := runtime.CallersFrames(pcs[:]).Next()
	line := strconv.Itoa(f.Line)
	for _, test := range handlerTests {
		r := slog.NewRecord(testTime, LevelInfo, "message", pcs[0])
		r.AddAttrs(test.attrs...)
		var buf bytes.Buffer
		var h slog.Handler = NewNanoHandler(&buf, Options{LevelInfo, false, test.addSource})
		t.Run(test.name, func(t *testing.T) {
			if test.preAttrs != nil {
				h = h.WithAttrs(test.preAttrs)
			}
			buf.Reset()
			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatalf("Handle() error: %v", err)
			}
			got := strings.TrimSuffix(buf.String(), "\n")
			want := strings.ReplaceAll(test.wantNano, "$LINE", line)
			if got != want {
				t.Errorf("\ngot  %s\nwant %s\n", got, want)
			}
		})
	}
}

func TestNanoHandlerRace(t *testing.T) {
	const P = 10
	const N = 10000
	done := make(chan struct{})
	h := NewNanoHandler(io.Discard, Options{LevelInfo, true, true})
	for i := range P {
		go func() {
			defer func() { done <- struct{}{} }()
			var pcs [1]uintptr
			runtime.Callers(1, pcs[:])
			r := slog.NewRecord(testTime, LevelInfo, "message", pcs[0])
			for j := range N {
				if err := h.Handle(context.Background(), r); err != nil {
					t.Errorf("goroutine(%d.%d) direct Handle got error %v", i, j, err)
					return
				}
				h2 := h.WithAttrs([]slog.Attr{slog.Int("int", 123)})
				if err := h2.Handle(context.Background(), r); err != nil {
					t.Errorf("goroutine(%d.%d) with attrs Handle got error %v", i, j, err)
					return
				}
				h2 = h.WithGroup("group")
				h2 = h2.WithAttrs([]slog.Attr{slog.Int("int", 123)})
				if err := h2.Handle(context.Background(), r); err != nil {
					t.Errorf("goroutine(%d.%d) with group attrs Handle got error %v", i, j, err)
					return
				}
			}
		}()
	}
	for range P {
		<-done
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
		{slog.AnyValue(errors.New("io error")), " io error"},
		{slog.AnyValue(io.EOF), " EOF"},
		{slog.AnyValue([]byte("test")), " [116 101 115 116]"},
		{slog.AnyValue(map[string]int{"age": 18}), " map[age:18]"},
		{slog.AnyValue(AnsiString{ansi.RedFG, "test"}), " test"},
		{slog.AnyValue(AnsiString{ansi.RedFG, "test"}), " \x1b[31mtest\x1b[0m"},
		{slog.AnyValue(errors.New("io error")), " \x1b[31mio error\x1b[0m"},
		{slog.AnyValue(io.EOF), " \x1b[31mEOF\x1b[0m"},
	}
	buf := make([]byte, 32)
	for i, test := range tests {
		buf = buf[:0]
		appendNanoValue(&buf, test.input, i >= len(tests)-3)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendNanoValue(%+v) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestAppendNanoSource(t *testing.T) {
	var tests = []struct {
		depth int
		re    string
	}{
		{0, `^runtime/extern.go:\d+$`},
		{1, `^logger/nano_handler_test.go:\d+$`},
		{2, `^testing/testing.go:\d+$`},
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
		if !regexp.MustCompile(test.re).MatchString(got) {
			t.Fatalf("appendNanoSource(%d) = %s, want matched by %s", test.depth, got, test.re)
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
