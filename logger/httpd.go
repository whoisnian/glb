package logger

import (
	"context"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/whoisnian/glb/httpd"
	"github.com/whoisnian/glb/util/strutil"
)

func (l *Logger) Relay(store *httpd.Store) {
	start := time.Now()
	remoteAddr := store.R.RemoteAddr[0:strings.LastIndexByte(store.R.RemoteAddr, ':')]
	if remoteAddr[0] == '[' {
		remoteAddr = remoteAddr[1 : len(remoteAddr)-1]
	}
	if l.h.Enabled(LevelInfo) {
		r := slog.NewRecord(time.Now(), LevelInfo, "", 0)
		r.AddAttrs(
			slog.String("tag", "REQ_BEG"),
			slog.String("ip", remoteAddr),
			slog.String("method", store.R.Method),
			slog.String("path", store.R.RequestURI),
			slog.String("tid", store.GetID()),
		)
		l.h.Handle(context.Background(), r)
	}
	defer func() {
		if l.h.Enabled(LevelInfo) {
			r := slog.NewRecord(time.Now(), LevelInfo, "", 0)
			r.AddAttrs(
				slog.String("tag", "REQ_END"),
				slog.Int("code", store.W.Status),
				slog.Int64("dur", time.Since(start).Milliseconds()),
				slog.String("ip", remoteAddr),
				slog.String("method", store.R.Method),
				slog.String("path", store.R.RequestURI),
				slog.String("tid", store.GetID()),
			)
			l.h.Handle(context.Background(), r)
		}
	}()
	defer func() {
		// https://cs.opensource.google/go/go/+/refs/tags/go1.21.1:src/net/http/server.go;l=1865
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			if l.h.Enabled(LevelError) {
				r := slog.NewRecord(time.Now(), LevelError, strutil.UnsafeBytesToString(buf), 0)
				r.AddAttrs(slog.Any("panic", err))
				r.AddAttrs(slog.String("tid", store.GetID()))
				l.h.Handle(context.Background(), r)
			}
			if store.W.Status == 0 {
				http.Error(store.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}()

	store.I.HandlerFunc(store)
}
