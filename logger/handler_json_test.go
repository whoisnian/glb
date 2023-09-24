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

func TestJsonHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewJsonHandler(&buf, NewOptions(LevelInfo, false, false))

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
	got, want := buf.String(), `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"m","dur":60000000000,"b":true,"a":1}`+"\n"
	if got != want {
		t.Errorf("new.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"m","a":1}`+"\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestJsonHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	var h Handler = NewJsonHandler(&buf, NewOptions(LevelInfo, false, false))

	hh := h.WithGroup("s")
	r := slog.NewRecord(testTime, LevelInfo, "m", 0)
	r.AddAttrs(slog.Duration("dur", time.Millisecond))
	r.AddAttrs(slog.Int("a", 2))

	// new1 Handler
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new1.Handle() error: %v", err)
	}
	got, want := buf.String(), `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"m","s":{"dur":1000000,"a":2}}`+"\n"
	if got != want {
		t.Errorf("new1.Handle() got %q, want %q", got, want)
	}

	// new2 Handler
	hh = hh.WithGroup("t")
	buf.Reset()
	if err := hh.Handle(context.Background(), r); err != nil {
		t.Fatalf("new2.Handle() error: %v", err)
	}
	got, want = buf.String(), `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"m","s":{"t":{"dur":1000000,"a":2}}}`+"\n"
	if got != want {
		t.Errorf("new2.Handle() got %q, want %q", got, want)
	}

	// origin Handler
	buf.Reset()
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("origin.Handle() error: %v", err)
	}
	got, want = buf.String(), `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"m","dur":1000000,"a":2}`+"\n"
	if got != want {
		t.Errorf("origin.Handle() got %q, want %q", got, want)
	}
}

func TestJsonHandler(t *testing.T) {
}

func TestJsonHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := 0; i < 10; i++ {
		r.AddAttrs(slog.Int("x", i))
	}
	var h Handler = NewJsonHandler(io.Discard, NewOptions(LevelInfo, false, false))
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

// as json.Marshaler for TestAppendJsonAttr
type jsonM struct{ s string }

func (t jsonM) MarshalJSON() ([]byte, error) {
	if len(t.s) == 0 {
		return nil, errors.New("EMPTY")
	} else if len(t.s) == 1 {
		return []byte(t.s), nil
	}
	return []byte(fmt.Sprintf("\"JSON{%s}\"", t.s)), nil
}

func TestAppendJsonAttr(t *testing.T) {
	var tests = []struct {
		input slog.Attr
		want  string
	}{
		{slog.String("a", "b"), `"a":"b"`},
		{slog.String("a b", "c d"), `"a b":"c d"`},
		{slog.Int("x", 123), `"x":123`},
		{slog.Int64("y", -456), `"y":-456`},
		{slog.Int64("y", -9007199254740991), `"y":-9007199254740991`},
		{slog.Uint64("z", 789), `"z":789`},
		{slog.Uint64("z", 9007199254740991), `"z":9007199254740991`},
		{slog.Float64("f", 0.125), `"f":0.125`},
		{slog.Float64("f", 1e-7), `"f":1e-7`},
		{slog.Bool("ok", false), `"ok":false`},
		{slog.Bool("ok", true), `"ok":true`},
		{slog.Time("tm", testTime), `"tm":"2000-01-02T03:04:05.000000006Z"`},
		{slog.Duration("dur", time.Minute), `"dur":60000000000`},
		{slog.Duration("dur", time.Microsecond), `"dur":1000`},
		{slog.Group("req",
			slog.String("method", "GET"),
			slog.Int("code", 200),
		), `"req":{"method":"GET","code":200}`},
		{slog.Any("data", []byte("test")), `"data":"test"`},
		{slog.Any("data", []byte{0x00, 0x01}), `"data":"\u0000\u0001"`},
		{slog.Any("map", map[string]int{"age": 18}), `"map":"map[age:18]"`},
		{slog.Any("map", map[string]string{"name": "a b"}), `"map":"map[name:a b]"`},
		{slog.Any("e", errors.New("io error")), `"e":"io error"`},
		{slog.Any("e", io.EOF), `"e":"EOF"`},
		{slog.Any("t", jsonM{""}), `"t":"EMPTY"`},
		{slog.Any("t", jsonM{"E"}), `"t":"invalid character 'E' looking for beginning of value"`},
		{slog.Any("t", jsonM{"value"}), `"t":"JSON{value}"`},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		buf = buf[:0]
		appendJsonAttr(&buf, test.input, false)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendJsonAttr(%+v) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestAppendJsonSource(t *testing.T) {
	var tests = []struct {
		depth  int
		prefix string
	}{
		{0, `"file":"runtime/extern.go","line":`},
		{1, `"file":"logger/handler_json_test.go","line":`},
		{2, `"file":"testing/testing.go","line":`},
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
		appendJsonSource(&buf, pc)

		got := string(buf)
		if !strings.HasPrefix(got, test.prefix) || !regexp.MustCompile(`\d+`).MatchString(got[len(test.prefix):]) {
			t.Fatalf("appendJsonSource(%d) = %s, want %sxxx", test.depth, got, test.prefix)
		}
	}
}

func TestAppendJsonString(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"", ``},
		{"ab", `ab`},
		{"a b", `a b`},
		{`"ab"`, `\"ab\"`},
		{"a=b", `a=b`},
		{"-123", `-123`},
		{"\r\n\t\a", `\r\n\t\u0007`},
		{"a\tb", `a\tb`},
		{`"[{escape}]"`, `\"[{escape}]\"`},
		{"<escapeHTML&>", `<escapeHTML&>`},
		{"\u03B8\u2028\u2029\uFFFF\xF6", `θ\u2028\u2029￿\ufffd`},
		{"µåπ", `µåπ`},
		{"badutf8\xF6", `badutf8\ufffd`},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		buf = buf[:0]
		appendJsonString(&buf, test.input)

		got := string(buf)
		if got != test.want {
			t.Fatalf("appendJsonString(%q) = %s, want %s", test.input, got, test.want)
		}
	}
}