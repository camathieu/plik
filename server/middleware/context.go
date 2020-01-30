package middleware

import (
	"github.com/root-gg/plik/server/context"
	"net/http"
)

// Log the http request
func Context(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx.SetRequest(req)
		ctx.SetResponse(resp)

		next.ServeHTTP(resp, req)
	})
}
