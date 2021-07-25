package httpd

import (
	"encoding/json"
	"net/http"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Store consists of responseWriter, request and routeParams.
type Store struct {
	W *statusResponseWriter
	R *http.Request
	m map[string]string
}

type Handler func(Store)

// CreateHandler converts 'http.HandlerFunc' to 'httpd.Handler'.
func CreateHandler(httpHandler http.HandlerFunc) Handler {
	return func(store Store) { httpHandler(store.W, store.R) }
}

// RouteParam returns the value of specified route param, or empty string if param not found.
func (store Store) RouteParam(name string) string {
	if param, ok := store.m[name]; ok {
		return param
	}
	return ""
}

// RouteParamAny returns the value of route param "/*".
func (store Store) RouteParamAny() string {
	return store.RouteParam(routeParamAny)
}

// CookieValue returns the value of specified cookie, or empty string if cookie not found.
func (store Store) CookieValue(name string) string {
	if cookie, err := store.R.Cookie(name); err == nil {
		return cookie.Value
	}
	return ""
}

// Respond200 replies 200 to client request with optional body.
func (store Store) Respond200(content []byte) error {
	store.W.WriteHeader(http.StatusOK)
	if len(content) > 0 {
		_, err := store.W.Write(content)
		return err
	}
	return nil
}

// RespondJson replies json body to client request.
func (store Store) RespondJson(v interface{}) error {
	store.W.Header().Add("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(store.W).Encode(v)
}

// Redirect is similar to `http.Redirect()`.
func (store Store) Redirect(url string, code int) {
	http.Redirect(store.W, store.R, url, code)
}

// Error404 is similar to `http.Error()`.
func (store Store) Error404(err string) {
	http.Error(store.W, err, http.StatusNotFound)
}

// Error500 is similar to `http.Error()`.
func (store Store) Error500(err string) {
	http.Error(store.W, err, http.StatusInternalServerError)
}
