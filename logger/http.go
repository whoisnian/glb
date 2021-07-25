package logger

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

// Example:
//   if err := http.ListenAndServe(":8080", logger.Req(http.DefaultServeMux)); err != nil {
//       logger.Fatal(err)
//   }
func Req(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggerResponseWriter{w, 0, time.Now()}
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
