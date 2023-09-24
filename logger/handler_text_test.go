package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTextHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewTextHandler(&buf, NewOptions(LevelInfo, false, false))

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
	got, want := buf.String(), "time=2000-01-02T03:04:05Z level=INFO msg=m dur=1m0s b=true a=1\n"
	if got != want {
		t.Errorf("new.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), "time=2000-01-02T03:04:05Z level=INFO msg=m a=1\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestTextHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewTextHandler(&buf, NewOptions(LevelInfo, false, false))

	hh := h.WithGroup("s")
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Duration("dur", time.Millisecond))
	r.AddAttrs(slog.Int("a", 2))

	// new Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new.Handle() error: %v", err)
	}
	got, want := buf.String(), "time=2000-01-02T03:04:05Z level=INFO msg=m s.dur=1ms s.a=2\n"
	if got != want {
		t.Errorf("new.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), "time=2000-01-02T03:04:05Z level=INFO msg=m dur=1ms a=2\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestTextHandler(t *testing.T) {
}

func TestTextHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := 0; i < 10; i++ {
		r.AddAttrs(slog.Int("x = y", i))
	}
	var h Handler = NewTextHandler(io.Discard, NewOptions(LevelInfo, false, false))
	got := int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("origin.Handle() got %d allocs, want 0", got)
	}

	h = h.WithGroup("s")
	r.AddAttrs(slog.Group("g", slog.Int64("a", 1)))
	got = int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("new.Handle() got %d allocs, want 0", got)
	}
}

// as TextMarshaler for TestAppendTextAttr
type textM struct{ s string }

func (t textM) MarshalText() ([]byte, error) {
	if len(t.s) == 0 {
		return nil, errors.New("EMPTY")
	}
	return []byte(fmt.Sprintf("TEXT{%s}", t.s)), nil
}

func TestAppendTextAttr(t *testing.T) {
	var tests = []struct {
		input slog.Attr
		want  string
	}{
		{slog.String("a", "b"), ` a=b`},
		{slog.String("a b", "c d"), ` "a b"="c d"`},
		{slog.Int("x", 123), ` x=123`},
		{slog.Int64("y", -456), ` y=-456`},
		{slog.Uint64("z", 789), ` z=789`},
		{slog.Float64("f", 0.125), ` f=0.125`},
		{slog.Float64("f", 1e-7), ` f=1e-07`},
		{slog.Bool("ok", false), ` ok=false`},
		{slog.Bool("ok", true), ` ok=true`},
		{slog.Time("tm", testTime), ` tm=2000-01-02T03:04:05Z`},
		{slog.Duration("dur", time.Minute), ` dur=1m0s`},
		{slog.Duration("dur", time.Microsecond), ` dur=1µs`},
		{slog.Group("req",
			slog.String("method", "GET"),
			slog.Int("code", 200),
		), ` req.method=GET req.code=200`},
		{slog.Any("data", []byte("test")), ` data=test`},
		{slog.Any("data", []byte{0x00, 0x01}), ` data="\x00\x01"`},
		{slog.Any("map", map[string]int{"age": 18}), ` map=map[age:18]`},
		{slog.Any("map", map[string]string{"name": "a b"}), ` map="map[name:a b]"`},
		{slog.Any("e", errors.New("io error")), ` e="io error"`},
		{slog.Any("e", io.EOF), ` e=EOF`},
		{slog.Any("t", textM{""}), ` t=EMPTY`},
		{slog.Any("t", textM{"value"}), ` t=TEXT{value}`},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		buf = buf[:0]
		appendTextAttr(&buf, test.input, "")

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendTextAttr(%+v) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestAppendTextSource(t *testing.T) {
	var tests = []struct {
		depth  int
		prefix string
	}{
		{0, "runtime/extern.go:"},
		{1, "logger/handler_text_test.go:"},
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
		appendTextSource(&buf, pc)

		got := string(buf)
		if !strings.HasPrefix(got, test.prefix) || !regexp.MustCompile(`\d+`).MatchString(got[len(test.prefix):]) {
			t.Fatalf("appendTextSource(%d) = %s, want %sxxx", test.depth, got, test.prefix)
		}
	}
}

func TestAppendTextString(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"", `""`},
		{"ab", `ab`},
		{"a b", `"a b"`},
		{`"ab"`, `"\"ab\""`},
		{"a=b", `"a=b"`},
		{"\a\b", `"\a\b"`},
		{"a\tb", `"a\tb"`},
		{"µåπ", `µåπ`},
		{"badutf8\xF6", `"badutf8\xf6"`},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		buf = buf[:0]
		appendTextString(&buf, test.input)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendTextString(%q) = %s, want %s", test.input, got, test.want)
		}
	}
}
