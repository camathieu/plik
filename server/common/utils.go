package common

import (
	"net/http"
	"strings"
)

// Ensure HTTPError implements error
var _ error = (*HTTPError)(nil)

// HTTPError allows to return an error and a HTTP status code
type HTTPError struct {
	error      string
	statusCode int
}

// NewHTTPError return a new HTTPError
func NewHTTPError(message string, code int) HTTPError {
	return HTTPError{message, code}
}

// Error return the error
func (e HTTPError) Error() string {
	return e.error
}

// GetStatusCode return the http status code
func (e HTTPError) GetStatusCode() int {
	return e.statusCode
}

// StripPrefix returns a handler that serves HTTP requests
// removing the given prefix from the request URL's Path
// It differs from http.StripPrefix by defaulting to "/" and not ""
func StripPrefix(prefix string, handler http.Handler) http.Handler {
	if prefix == "" || prefix == "/" {
		return handler
	}
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// Relative paths to javascript, css, ... imports won't work without a tailing slash
		if req.URL.Path == prefix {
			http.Redirect(resp, req, prefix+"/", http.StatusMovedPermanently)
			return
		}
		if p := strings.TrimPrefix(req.URL.Path, prefix); len(p) < len(req.URL.Path) {
			req.URL.Path = p
		} else {
			http.NotFound(resp, req)
			return
		}
		if !strings.HasPrefix(req.URL.Path, "/") {
			req.URL.Path = "/" + req.URL.Path
		}
		handler.ServeHTTP(resp, req)
	})
}
