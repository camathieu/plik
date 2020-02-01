package middleware

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/root-gg/plik/server/context"
)

// Yubikey verify that a valid OTP token has been provided
func Yubikey(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		config := ctx.GetConfig()

		// Get upload from context
		upload := ctx.GetUpload()
		if upload == nil {
			ctx.InternalServerError("missing upload in yubikey middleware", nil)
			return
		}

		// If upload is yubikey protected, user must send an OTP when he wants to get a file.
		if upload.Yubikey != "" {

			// Error if yubikey is disabled on server, and enabled on upload
			if !config.YubikeyEnabled {
				ctx.BadRequest("yubikey are disabled on this server")
				return
			}

			vars := mux.Vars(req)
			token := vars["yubikey"]
			if token == "" {
				ctx.BadRequest("missing yubikey token")
				return
			}
			if len(token) != 44 {
				ctx.BadRequest("invalid yubikey token")
				return
			}
			if token[:12] != upload.Yubikey {
				ctx.BadRequest("invalid yubikey token")
				return
			}

			_, isValid, err := config.GetYubiAuth().Verify(token)
			if err != nil {
				ctx.InternalServerError("unable to validate yubikey token", err)
				return
			}
			if !isValid {
				ctx.Unauthorized("invalid yubikey token")
				return
			}
		}

		next.ServeHTTP(resp, req)
	})
}
