// Package logger implements a simple logger based on log/slog.
//
// NanoHandler formats slog.Record as a sequence of value strings without attribute keys to minimize log length.
//
//	2023/08/16 00:35:15 [I] Service httpd started: <http://127.0.0.1:9000>
//	2023/08/16 00:35:15 [D] Use default cache dir: /root/.cache/glb
//	2023/08/16 00:35:15 [W] Failed to check for updates: i/o timeout
//	2023/08/16 00:35:15 [E] dial tcp 127.0.0.1:9000: connect: connection refused
//	2023/08/16 00:35:15 [F] mkdir /root/.config/glb: permission denied
//
//	2023/08/16 00:35:15 [I] REQUEST 10.0.3.201 GET /status R5U3KA5C-42
//	2023/08/16 00:35:15 [I] FETCH Fetch upstream metrics http://127.0.0.1:8080/metrics 3 R5U3KA5C-42
//	2023/08/16 00:35:15 [I] RESPONSE 200 4 10.0.3.201 GET /status R5U3KA5C-42
//
// TextHandler formats slog.Record as a sequence of key=value pairs separated by spaces and followed by a newline.
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Service httpd started: <http://127.0.0.1:9000>"
//	time=2023-08-16T00:35:15+08:00 level=DEBUG msg="Use default cache dir: /root/.cache/glb"
//	time=2023-08-16T00:35:15+08:00 level=WARN msg="Failed to check for updates: i/o timeout"
//	time=2023-08-16T00:35:15+08:00 level=ERROR msg="dial tcp 127.0.0.1:9000: connect: connection refused"
//	time=2023-08-16T00:35:15+08:00 level=FATAL msg="mkdir /root/.config/glb: permission denied"
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" tag=REQUEST ip=10.0.3.201 method=GET path=/status tid=R5U3KA5C-42
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Fetch upstream metrics" tag=FETCH url=http://127.0.0.1:8080/metrics duration=3 tid=R5U3KA5C-42
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" tag=RESPONSE code=200 duration=4 ip=10.0.3.201 method=GET path=/status tid=R5U3KA5C-42
//
// JsonHandler formats slog.Record as line-delimited JSON objects.
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Service httpd started: <http://127.0.0.1:9000>"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"DEBUG","msg":"Use default cache dir: /root/.cache/glb"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"WARN","msg":"Failed to check for updates: i/o timeout"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"ERROR","msg":"dial tcp 127.0.0.1:9000: connect: connection refused"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"FATAL","msg":"mkdir /root/.config/glb: permission denied"}
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","tag":"REQUEST","ip":"10.0.3.201","method":"GET","path":"/status","tid":"R5U3KA5C-42"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Fetch upstream metrics","tag":"FETCH","url":"http://127.0.0.1:8080/metrics","duration":3,"tid":"R5U3KA5C-42"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","tag":"RESPONSE","code":200,"duration":4,"ip":"10.0.3.201","method":"GET","path":"/status","tid":"R5U3KA5C-42"}
package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	h Handler
}

func New(h Handler) *Logger {
	return &Logger{h}
}

func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}
	return &Logger{l.h.WithAttrs(argsToAttrs(args))}
}

func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}
	return &Logger{l.h.WithGroup(name)}
}

func argsToAttrs(args []any) (attrs []slog.Attr) {
	const badKey = "!BADKEY"

	for i := 0; i < len(args); i++ {
		switch x := args[i].(type) {
		case string:
			if i+1 < len(args) {
				attrs = append(attrs, slog.Any(x, args[i+1]))
				i++
			} else {
				attrs = append(attrs, slog.String(badKey, x))
			}
		case slog.Attr:
			attrs = append(attrs, x)
		default:
			attrs = append(attrs, slog.Any(badKey, x))
		}
	}
	return attrs
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), LevelDebug, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), LevelError, msg, args...)
}

func (l *Logger) Panic(msg string, args ...any) {
	l.log(context.Background(), LevelError, msg, args...)
	panic(msg)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) error {
	if !l.h.Enabled(level) {
		return nil
	}
	var pc uintptr
	if l.h.IsAddSource() {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	return l.h.Handle(ctx, r)
}

func (l *Logger) logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) error {
	if !l.h.Enabled(level) {
		return nil
	}
	var pc uintptr
	if l.h.IsAddSource() {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	return l.h.Handle(ctx, r)
}
