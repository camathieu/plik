package middleware

import (
	"fmt"
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
			ctx.InternalServerError(fmt.Errorf("missing upload in yubikey middleware"))
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
				ctx.Unauthorized("missing yubikey token")
				return
			}
			if len(token) != 44 {
				ctx.Unauthorized("invalid yubikey token")
				return
			}
			if token[:12] != upload.Yubikey {
				ctx.Unauthorized("invalid yubikey token")
				return
			}

			_, isValid, err := config.GetYubiAuth().Verify(token)
			if err != nil {
				ctx.InternalServerError(fmt.Errorf("unable to validate yubikey token : %s", err))
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
