package httpd

import (
	"encoding/json"
	"net/http"
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

// Store consists of responseWriter, request and routeParams.
type Store struct {
	W http.ResponseWriter
	R *http.Request
	P *Params
}

type HandlerFunc func(*Store)

// CreateHandler converts 'http.HandlerFunc' to 'httpd.HandlerFunc'.
func CreateHandler(httpHandler http.HandlerFunc) HandlerFunc {
	return func(store *Store) { httpHandler(store.W, store.R) }
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

// RespondJson replies json body to client request.
func (store *Store) RespondJson(v interface{}) error {
	store.W.Header().Add("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(store.W).Encode(v)
}

// Redirect is similar to `http.Redirect()`.
func (store *Store) Redirect(url string, code int) {
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
