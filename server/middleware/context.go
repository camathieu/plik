package middleware

import (
	"net/http"

	"github.com/root-gg/plik/server/context"
)

// Log the http request
func Context(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx.SetReq(req)
		ctx.SetResp(resp)

		next.ServeHTTP(resp, req)
	})
}
