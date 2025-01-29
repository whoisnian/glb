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
//	2023/08/16 00:35:15 [I] REQ_BEG 10.0.3.201 GET /status
//	2023/08/16 00:35:15 [I] Fetch upstream http://127.0.0.1:8080/metrics, duration 3ms
//	2023/08/16 00:35:15 [I] REQ_END 200 4 10.0.3.201 GET /status
//
// TextHandler formats slog.Record as a sequence of key=value pairs separated by spaces and followed by a newline.
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Service httpd started: <http://127.0.0.1:9000>"
//	time=2023-08-16T00:35:15+08:00 level=DEBUG msg="Use default cache dir: /root/.cache/glb"
//	time=2023-08-16T00:35:15+08:00 level=WARN msg="Failed to check for updates: i/o timeout"
//	time=2023-08-16T00:35:15+08:00 level=ERROR msg="dial tcp 127.0.0.1:9000: connect: connection refused"
//	time=2023-08-16T00:35:15+08:00 level=FATAL msg="mkdir /root/.config/glb: permission denied"
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" request.tag=REQ_BEG request.ip=10.0.3.201 request.method=GET request.path=/status
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Fetch upstream http://127.0.0.1:8080/metrics, duration 3ms"
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" request.tag=REQ_END request.code=200 request.dur=4 request.ip=10.0.3.201 request.method=GET request.path=/status
//
// JsonHandler formats slog.Record as line-delimited JSON objects.
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Service httpd started: <http://127.0.0.1:9000>"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"DEBUG","msg":"Use default cache dir: /root/.cache/glb"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"WARN","msg":"Failed to check for updates: i/o timeout"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"ERROR","msg":"dial tcp 127.0.0.1:9000: connect: connection refused"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"FATAL","msg":"mkdir /root/.config/glb: permission denied"}
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","request":{"tag":"REQ_BEG","ip":"10.0.3.201","method":"GET","path":"/status"}}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Fetch upstream http://127.0.0.1:8080/metrics, duration 3ms"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","request":{"tag":"REQ_END","code":200,"dur":4,"ip":"10.0.3.201","method":"GET","path":"/status"}}
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Logger provides output methods for creating a [slog.Record] and passing it to the internal [Handler].
type Logger struct {
	addSource bool
	handler   slog.Handler
}

// New creates a new Logger with the given Handler.
func New(h slog.Handler) *Logger {
	return &Logger{tryIsAddSource(h), h}
}

// With returns a Logger that includes the given args in each output operation.
// If args is empty, With returns the origin Logger.
func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}
	return &Logger{l.addSource, l.handler.WithAttrs(argsToAttrs(args))}
}

// WithGroup returns a Logger that starts a group with the given name.
// If name is empty, WithGroup returns the origin Logger.
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}
	return &Logger{l.addSource, l.handler.WithGroup(name)}
}

// argsToAttrs is equivalent to slog.argsToAttrSlice().
//
//   - If args[i] is an Attr, it is used as is.
//   - If args[i] is a string and is not the last argument, it is treated as slog.Attr{args[i], args[i+1]}.
//   - Otherwise, the args[i] is treated as slog.Attr{"!BADKEY", args[i]}.
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

// Enabled reports whether the given level is enabled.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Enabled(ctx, level)
}

// Debug logs at LevelDebug.
//
// If args need heavy computation, should wrap them with a conditional block or
// use the LogValuer interface to defer expensive operations. Example:
//
//	if logger.Enabled(ctx, LevelDebug) {
//	    result := calcDifference(A, B)
//	    logger.Debug(ctx, "difference between A and B: " + result)
//	}
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

// Debugf formats message at LevelDebug.
func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, LevelDebug, format, args...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

// Infof formats message at LevelInfo.
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.logf(ctx, LevelInfo, format, args...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

// Warnf formats message at LevelWarn.
func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, LevelWarn, format, args...)
}

// Error logs at LevelError.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

// Errorf formats message at LevelError.
func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, LevelError, format, args...)
}

// Fatal logs at LevelFatal and follows with a call to os.Exit(1).
func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// Fatalf formats message at LevelFatal and follows with a call to os.Exit(1).
func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, LevelFatal, format, args...)
	os.Exit(1)
}

// Log emits a log record with the current time and the given level and message.
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// Logf emits a log record with the current time and the given level and format message.
func (l *Logger) Logf(ctx context.Context, level slog.Level, format string, args ...any) {
	l.logf(ctx, level, format, args...)
}

// LogAttrs is a more efficient version of [Logger.Log] that accepts only Attrs.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) error {
	if !l.Enabled(ctx, level) {
		return nil
	}
	var pc uintptr
	if l.addSource {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Handle(ctx, r)
}

func (l *Logger) logf(ctx context.Context, level slog.Level, format string, args ...any) error {
	if !l.Enabled(ctx, level) {
		return nil
	}
	var pc uintptr
	if l.addSource {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pc)
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Handle(ctx, r)
}

func (l *Logger) logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) error {
	if !l.Enabled(ctx, level) {
		return nil
	}
	var pc uintptr
	if l.addSource {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Handle(ctx, r)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

type AnsiString struct {
	Prefix string
	Value  string
}
