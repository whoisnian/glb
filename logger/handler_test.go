package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestOptionsEnabled(t *testing.T) {
	var tests = []struct {
		opts  Options
		input slog.Level
		want  bool
	}{
		{Options{level: LevelDebug}, LevelDebug, true},
		{Options{level: LevelDebug}, LevelInfo, true},
		{Options{level: LevelDebug}, LevelWarn, true},
		{Options{level: LevelInfo}, LevelDebug, false},
		{Options{level: LevelInfo}, LevelInfo, true},
		{Options{level: LevelInfo}, LevelError, true},
		{Options{level: LevelError}, LevelDebug, false},
		{Options{level: LevelError}, LevelWarn, false},
		{Options{level: LevelError}, LevelError, true},
		{Options{level: LevelError}, LevelFatal, true},
	}
	for _, test := range tests {
		got := test.opts.Enabled(test.input)
		if got != test.want {
			t.Errorf("%v.Enabled(%d) = %v, want %v", test.opts, test.input, got, test.want)
		}
	}
}

func TestOptionsCheck(t *testing.T) {
	var tests = []struct {
		opts  *Options
		wantD bool
		wantC bool
		wantA bool
	}{
		{NewOptions(LevelDebug, false, false), true, false, false},
		{NewOptions(LevelDebug, true, false), true, true, false},
		{NewOptions(LevelInfo, true, true), false, true, true},
		{NewOptions(LevelWarn, false, false), false, false, false},
		{NewOptions(LevelError, false, true), false, false, true},
		{NewOptions(LevelFatal, true, true), false, true, true},
	}
	for _, test := range tests {
		if test.opts.IsDebug() != test.wantD {
			t.Errorf("%v.IsDebug() = %v, want %v", test.opts, test.opts.IsDebug(), test.wantD)
		}
		if test.opts.IsColorful() != test.wantC {
			t.Errorf("%v.IsColorful() = %v, want %v", test.opts, test.opts.IsColorful(), test.wantC)
		}
		if test.opts.IsAddSource() != test.wantA {
			t.Errorf("%v.IsAddSource() = %v, want %v", test.opts, test.opts.IsAddSource(), test.wantA)
		}
	}
}

// as slog.LogValuer for handlerTests
type slogLV struct {
	s string
}

func (lv slogLV) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("LOGVALUER{%s}", lv.s))
}

var (
	tmpAttrs    = []slog.Attr{slog.String("a", "one"), slog.Int("b", 2), slog.Duration("c", time.Millisecond)}
	tmpPreAttrs = []slog.Attr{slog.Int("pre", 3), slog.String("x", "y")}

	handlerTests = []struct {
		name      string
		addSource bool
		preAttrs  []slog.Attr
		attrs     []slog.Attr

		wantNano string
		wantText string
		wantJson string
	}{
		{
			name:     "basic",
			attrs:    tmpAttrs,
			wantNano: "2000-01-02 03:04:05 [I] message one 2 1ms",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message a=one b=2 c=1ms",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","a":"one","b":2,"c":1000000}`,
		}, {
			name:     "empty key",
			attrs:    []slog.Attr{slog.String("", "v")},
			wantNano: "2000-01-02 03:04:05 [I] message v",
			wantText: `time=2000-01-02T03:04:05Z level=INFO msg=message ""=v`,
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","":"v"}`,
		}, {
			name:     "preformatted",
			preAttrs: tmpPreAttrs,
			attrs:    tmpAttrs,
			wantNano: "2000-01-02 03:04:05 [I] message 3 y one 2 1ms",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message pre=3 x=y a=one b=2 c=1ms",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","pre":3,"x":"y","a":"one","b":2,"c":1000000}`,
		}, {
			name: "groups",
			attrs: []slog.Attr{
				slog.Int("a", 1),
				slog.Group("g",
					slog.Int("b", 2),
					slog.Group("h", slog.Int("c", 3)),
					slog.Int("d", 4)),
				slog.Int("e", 5),
			},
			wantNano: "2000-01-02 03:04:05 [I] message 1 2 3 4 5",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message a=1 g.b=2 g.h.c=3 g.d=4 e=5",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","a":1,"g":{"b":2,"h":{"c":3},"d":4},"e":5}`,
		}, {
			name:     "empty group",
			attrs:    []slog.Attr{slog.Group("g"), slog.Group("h", slog.Int("a", 1))},
			wantNano: "2000-01-02 03:04:05 [I] message 1",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message h.a=1",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","h":{"a":1}}`,
		}, {
			name: "nested empty group",
			attrs: []slog.Attr{
				slog.Group("g",
					slog.Group("h",
						slog.Group("i"),
						slog.Group("j"))),
			},
			wantNano: "2000-01-02 03:04:05 [I] message",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message"}`,
		}, {
			name: "nested non-empty group",
			attrs: []slog.Attr{
				slog.Group("g",
					slog.Group("h",
						slog.Group("i"),
						slog.Group("j", slog.Int("a", 1)))),
			},
			wantNano: "2000-01-02 03:04:05 [I] message 1",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message g.h.j.a=1",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","g":{"h":{"j":{"a":1}}}}`,
		}, {
			name: "escapes",
			attrs: []slog.Attr{
				slog.String("a b", "x\t\n\000y"),
				slog.Group(" b.c=\"\\x2E\t",
					slog.String("d=e", "f.g\""),
					slog.Int("m.d", 1)),
			},
			wantNano: "2000-01-02 03:04:05 [I] message x\t\n\000y f.g\" 1",
			wantText: `time=2000-01-02T03:04:05Z level=INFO msg=message "a b"="x\t\n\x00y" " b.c=\"\\x2E\t.d=e"="f.g\"" " b.c=\"\\x2E\t.m.d"=1`,
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","a b":"x\t\n\u0000y"," b.c=\"\\x2E\t":{"d=e":"f.g\"","m.d":1}}`,
		}, {
			name: "LogValuer",
			attrs: []slog.Attr{
				slog.Int("a", 1),
				slog.Any("name", slogLV{"value"}),
				slog.Int("b", 2),
			},
			wantNano: "2000-01-02 03:04:05 [I] message 1 LOGVALUER{value} 2",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message a=1 name=LOGVALUER{value} b=2",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","a":1,"name":"LOGVALUER{value}","b":2}`,
		}, {
			name:     "byte slice",
			attrs:    []slog.Attr{slog.Any("bs", []byte{1, 2, 3, 4})},
			wantNano: "2000-01-02 03:04:05 [I] message [1 2 3 4]",
			wantText: `time=2000-01-02T03:04:05Z level=INFO msg=message bs="\x01\x02\x03\x04"`,
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","bs":"AQIDBA=="}`,
		}, {
			name:     "json.RawMessage",
			attrs:    []slog.Attr{slog.Any("bs", json.RawMessage([]byte("1234")))},
			wantNano: "2000-01-02 03:04:05 [I] message [49 50 51 52]",
			wantText: `time=2000-01-02T03:04:05Z level=INFO msg=message bs="[49 50 51 52]"`,
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","bs":1234}`,
		}, {
			name: "inline group",
			attrs: []slog.Attr{
				slog.Int("a", 1),
				slog.Group("", slog.Int("b", 2), slog.Int("c", 3)),
				slog.Int("d", 4),
			},
			wantNano: "2000-01-02 03:04:05 [I] message 1 2 3 4",
			wantText: "time=2000-01-02T03:04:05Z level=INFO msg=message a=1 b=2 c=3 d=4",
			wantJson: `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","msg":"message","a":1,"b":2,"c":3,"d":4}`,
		}, {
			name:      "Source",
			addSource: true,
			wantNano:  "2000-01-02 03:04:05 [I] logger/nano_handler_test.go:$LINE message",
			wantText:  "time=2000-01-02T03:04:05Z level=INFO source=logger/text_handler_test.go:$LINE msg=message",
			wantJson:  `{"time":"2000-01-02T03:04:05.000000006Z","level":"INFO","source":{"file":"logger/json_handler_test.go","line":$LINE},"msg":"message"}`,
		},
	}
)
