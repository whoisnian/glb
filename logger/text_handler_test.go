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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/whoisnian/glb/ansi"
)

func TestTextHandlerEnabled(t *testing.T) {
	for _, test := range levelTests {
		got := NewTextHandler(io.Discard, Options{Level: test.opts}).Enabled(context.Background(), test.input)
		if got != test.want {
			t.Errorf("(%v).Enabled(ctx, %d) = %v, want %v", test.opts, test.input, got, test.want)
		}
	}
}

func TestTextHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h slog.Handler = NewTextHandler(&buf, Options{LevelInfo, false, false})

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
	var h slog.Handler = NewTextHandler(&buf, Options{LevelInfo, false, false})

	hh := h.WithGroup("s")
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Duration("dur", time.Millisecond))
	r.AddAttrs(slog.Int("a", 2))

	// new1 Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new1.Handle() error: %v", err)
	}
	got, want := buf.String(), "time=2000-01-02T03:04:05Z level=INFO msg=m s.dur=1ms s.a=2\n"
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
	got, want = buf.String(), `time=2000-01-02T03:04:05Z level=INFO msg=m s.b=true s.c=3 s.dur=1ms s.a=2`+"\n"
	if got != want {
		t.Errorf("new2.Handle() got %q, want %q", got, want)
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
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])
	f, _ := runtime.CallersFrames(pcs[:]).Next()
	line := strconv.Itoa(f.Line)
	for _, test := range handlerTests {
		r := slog.NewRecord(testTime, LevelInfo, "message", pcs[0])
		r.AddAttrs(test.attrs...)
		var buf bytes.Buffer
		var h slog.Handler = NewTextHandler(&buf, Options{LevelInfo, false, test.addSource})
		t.Run(test.name, func(t *testing.T) {
			if test.preAttrs != nil {
				h = h.WithAttrs(test.preAttrs)
			}
			buf.Reset()
			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatalf("Handle() error: %v", err)
			}
			got := strings.TrimSuffix(buf.String(), "\n")
			want := strings.ReplaceAll(test.wantText, "$LINE", line)
			if got != want {
				t.Errorf("\ngot  %s\nwant %s\n", got, want)
			}
		})
	}
}

func TestTextHandlerRace(t *testing.T) {
	const P = 10
	const N = 10000
	done := make(chan struct{})
	h := NewTextHandler(io.Discard, Options{LevelInfo, true, true})
	for i := 0; i < P; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			var pcs [1]uintptr
			runtime.Callers(1, pcs[:])
			r := slog.NewRecord(testTime, LevelInfo, "message", pcs[0])
			for j := 0; j < N; j++ {
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
	for i := 0; i < P; i++ {
		<-done
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
		{slog.Any("as", AnsiString{ansi.RedFG, "test"}), " as=test"},
		{slog.Any("as", AnsiString{ansi.RedFG, "test"}), " as=\x1b[31mtest\x1b[0m"},
	}
	buf := make([]byte, 32)
	for i, test := range tests {
		buf = buf[:0]
		appendTextAttr(&buf, test.input, &[]byte{}, i == len(tests)-1)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendTextAttr(%+v) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestAppendTextSource(t *testing.T) {
	var tests = []struct {
		depth int
		re    string
	}{
		{0, `^runtime/extern.go:\d+$`},
		{1, `^logger/text_handler_test.go:\d+$`},
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
		appendTextSource(&buf, pc)

		got := string(buf)
		if !regexp.MustCompile(test.re).MatchString(got) {
			t.Fatalf("appendTextSource(%d) = %s, want matched by %s", test.depth, got, test.re)
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
