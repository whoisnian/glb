package logger

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"time"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/httpd"
)

func (l *Logger) NewMiddleware() httpd.HandlerFunc {
	return func(store *httpd.Store) {
		start := time.Now()
		clientIP := store.GetClientIP()
		if l.Enabled(store.R.Context(), LevelInfo) {
			r := slog.NewRecord(time.Now(), LevelInfo, "", 0)
			r.AddAttrs(slog.Attr{
				Key: "request",
				Value: slog.GroupValue(
					slog.String("tag", "REQ_BEG"),
					slog.String("ip", clientIP),
					slog.String("method", store.R.Method),
					slog.String("path", store.R.URL.Path),
					slog.String("query", store.R.URL.RawQuery),
				),
			})
			l.handler.Handle(store.R.Context(), r)
		}
		defer func() {
			if l.Enabled(store.R.Context(), LevelInfo) {
				if store.W.Status == 0 {
					store.W.Status = http.StatusOK
				}
				r := slog.NewRecord(time.Now(), LevelInfo, "", 0)
				r.AddAttrs(slog.Attr{
					Key: "request",
					Value: slog.GroupValue(
						slog.Any("tag", AnsiString{ansi.BlueFG, "REQ_END"}),
						slog.Int("code", store.W.Status),
						slog.Int64("dur", time.Since(start).Milliseconds()),
						slog.String("ip", clientIP),
						slog.String("method", store.R.Method),
						slog.String("path", store.R.URL.Path),
						slog.String("query", store.R.URL.RawQuery),
					),
				})
				l.handler.Handle(store.R.Context(), r)
			}
		}()
		defer func() {
			// https://cs.opensource.google/go/go/+/refs/tags/go1.23.5:src/net/http/server.go;l=1943
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				if l.Enabled(store.R.Context(), LevelError) {
					raw, _ := httputil.DumpRequest(store.R, false)
					r := slog.NewRecord(time.Now(), LevelError, "Recover from panic", 0)
					r.AddAttrs(slog.Attr{
						Key: "request",
						Value: slog.GroupValue(
							slog.Any("panic", err),
							slog.String("stack", string(debug.Stack())),
							slog.String("raw", string(raw)),
						),
					})
					l.handler.Handle(store.R.Context(), r)
				}
				if store.W.Status == 0 {
					http.Error(store.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}
		}()

		store.Next()
	}
}
