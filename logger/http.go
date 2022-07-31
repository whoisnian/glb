package logger

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// loggerResponseWriter records request start time and response status with http.ResponseWriter.
type loggerResponseWriter struct {
	w      http.ResponseWriter
	status int
	start  time.Time
}

func (lw *loggerResponseWriter) Header() http.Header {
	return lw.w.Header()
}

func (lw *loggerResponseWriter) Write(bytes []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	return lw.w.Write(bytes)
}

func (lw *loggerResponseWriter) WriteHeader(code int) {
	lw.status = code
	lw.w.WriteHeader(code)
}

func (lw *loggerResponseWriter) Flush() {
	if flusher, ok := lw.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Req wraps http.Handler and writes request log to stdout with tagR. Example:
//   if err := http.ListenAndServe(":8000", logger.Req(http.DefaultServeMux)); err != nil {
//       logger.Fatal(err)
//   }
func Req(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggerResponseWriter{w, 200, time.Now()}
		handler.ServeHTTP(lw, r)
		lout.Output(2, tagR+" "+fmt.Sprint(
			r.RemoteAddr[0:strings.IndexByte(r.RemoteAddr, ':')], " [",
			lw.status, "] ",
			r.Method, " ",
			r.RequestURI, " ",
			r.UserAgent(), " ",
			time.Since(lw.start).Milliseconds())+"\n")
	})
}

// Recovery wraps http.Handler and writes panic log to stderr with tagE. Example:
//   if err := http.ListenAndServe(":8000", logger.Recovery(http.DefaultServeMux)); err != nil {
//       logger.Fatal(err)
//   }
// If want use Recovery with Req, it should be like:
//   if err := http.ListenAndServe(":8000", logger.Req(logger.Recovery(http.DefaultServeMux))); err != nil {
//       logger.Fatal(err)
//   }
func Recovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/net/http/server.go;l=1822
			if err := recover(); err != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]

				lerr.Output(2, tagE+" panic: "+fmt.Sprint(err)+"\n"+string(buf))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}
