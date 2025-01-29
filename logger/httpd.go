package logger

import (
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/httpd"
	"github.com/whoisnian/glb/util/netutil"
	"github.com/whoisnian/glb/util/strutil"
)

func (l *Logger) NewMiddleware() httpd.HandlerFunc {
	return func(store *httpd.Store) {
		start := time.Now()
		remoteIP, _ := netutil.SplitHostPort(store.R.RemoteAddr)
		if l.Enabled(store.R.Context(), LevelInfo) {
			r := slog.NewRecord(time.Now(), LevelInfo, "", 0)
			r.AddAttrs(slog.Attr{
				Key: "request",
				Value: slog.GroupValue(
					slog.String("tag", "REQ_BEG"),
					slog.String("ip", remoteIP),
					slog.String("method", store.R.Method),
					slog.String("path", store.R.RequestURI),
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
						slog.String("ip", remoteIP),
						slog.String("method", store.R.Method),
						slog.String("path", store.R.RequestURI),
					),
				})
				l.handler.Handle(store.R.Context(), r)
			}
		}()
		defer func() {
			// https://cs.opensource.google/go/go/+/refs/tags/go1.23.5:src/net/http/server.go;l=1943
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				if l.Enabled(store.R.Context(), LevelError) {
					r := slog.NewRecord(time.Now(), LevelError, strutil.UnsafeBytesToString(buf), 0)
					r.AddAttrs(slog.Attr{
						Key:   "request",
						Value: slog.GroupValue(slog.Any("panic", err)),
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
