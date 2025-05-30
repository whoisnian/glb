package httpd

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/whoisnian/glb/util/netutil"
)

// Params consists of a pair of key-value slice.
type Params struct {
	K, V []string
}

// Get gets the first param value associated with the given key.
// The ok result indicates whether param was found.
func (ps *Params) Get(key string) (value string, ok bool) {
	for i := range ps.K {
		if ps.K[i] == key {
			return ps.V[i], true
		}
	}
	return "", false
}

// ResponseWriter records response status with http.ResponseWriter.
type ResponseWriter struct {
	Origin http.ResponseWriter
	Status int
}

func (w *ResponseWriter) Header() http.Header {
	return w.Origin.Header()
}

func (w *ResponseWriter) Write(bytes []byte) (int, error) {
	if w.Status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.Origin.Write(bytes)
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.Origin.WriteHeader(code)
	w.Status = code
}

// Flush implements the standard http.Flusher interface.
func (w *ResponseWriter) Flush() {
	if flusher, ok := w.Origin.(http.Flusher); ok {
		flusher.Flush()
	}
}

// FlushError attempts to invoke FlushError() of the standard http.ResponseWriter.
func (w *ResponseWriter) FlushError() error {
	if flusher, ok := w.Origin.(interface{ FlushError() error }); ok {
		return flusher.FlushError()
	} else if flusher, ok := w.Origin.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// Store consists of responseWriter, request, routeParams and routeInfo.
type Store struct {
	W *ResponseWriter
	R *http.Request
	P *Params
	I *RouteInfo

	mwIndex int
}

// Next should be used only in middleware to call the next middleware or handler.
func (store *Store) Next() {
	store.mwIndex++
	if store.mwIndex < len(*store.I.Middlewares) {
		(*store.I.Middlewares)[store.mwIndex](store)
	} else if store.mwIndex == len(*store.I.Middlewares) {
		store.I.HandlerFunc(store)
	}
}

type HandlerFunc func(*Store)

// CreateHandler converts 'http.HandlerFunc' to 'httpd.HandlerFunc'.
func CreateHandler(httpHandler http.HandlerFunc) HandlerFunc {
	return func(store *Store) { httpHandler(store.W, store.R) }
}

// GetClientIP looks for possible client IP by the following order:
//   - Header["X-Client-IP"]
//   - Header["X-Forwarded-For"]
//   - Header["X-Real-IP"]
//   - Request.RemoteAddr
//
// Warning: Malicious clients may use request header for IP spoofing.
func (store *Store) GetClientIP() string {
	if ip := store.R.Header.Get("X-Client-IP"); ip != "" {
		return ip
	}
	if ip := store.R.Header.Get("X-Forwarded-For"); ip != "" {
		if i := strings.IndexByte(ip, ','); i != -1 {
			return ip[:i]
		}
		return ip
	}
	if ip := store.R.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _ := netutil.SplitHostPort(store.R.RemoteAddr)
	return host
}

// RouteParam returns the value of specified route param, or empty string if param not found.
func (store *Store) RouteParam(name string) (value string) {
	value, _ = store.P.Get(name)
	return value
}

// RouteParamAny returns the value of route param "/*".
func (store *Store) RouteParamAny() string {
	return store.RouteParam(routeParamAny)
}

// CookieValue returns the value of specified cookie, or empty string if cookie not found.
func (store *Store) CookieValue(name string) string {
	if cookie, err := store.R.Cookie(name); err == nil {
		return cookie.Value
	}
	return ""
}

// Respond200 replies 200 to client request with optional body.
func (store *Store) Respond200(content []byte) error {
	store.W.WriteHeader(http.StatusOK)
	if len(content) > 0 {
		_, err := store.W.Write(content)
		return err
	}
	return nil
}

// RespondJson replies json body to client request. It cannot be called after `store.W.WriteHeader()`.
func (store *Store) RespondJson(code int, v any) error {
	store.W.Header().Add("Content-Type", "application/json; charset=utf-8")
	store.W.WriteHeader(code)
	return json.NewEncoder(store.W).Encode(v)
}

// Redirect is similar to `http.Redirect()`.
func (store *Store) Redirect(code int, url string) {
	http.Redirect(store.W, store.R, url, code)
}

// Error404 is similar to `http.Error()`.
func (store *Store) Error404(msg string) {
	http.Error(store.W, msg, http.StatusNotFound)
}

// Error500 is similar to `http.Error()`.
func (store *Store) Error500(msg string) {
	http.Error(store.W, msg, http.StatusInternalServerError)
}
