package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
)

// Yubikey verify that a valid OTP token has been provided
func Yubikey(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := context.GetLogger(ctx)
		config := context.GetConfig(ctx)

		// Get upload from context
		upload := context.GetUpload(ctx)
		if upload == nil {
			// This should never append
			log.Critical("Missing upload in yubikey middleware")
			context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
			return
		}

		// If upload is yubikey protected, user must send an OTP when he wants to get a file.
		if upload.Yubikey != "" {

			// Error if yubikey is disabled on server, and enabled on upload
			if !config.YubikeyEnabled {
				log.Warningf("Got a Yubikey upload but Yubikey backend is disabled")
				context.Fail(ctx, req, resp, "Yubikey are disabled on this server", http.StatusForbidden)
				return
			}

			vars := mux.Vars(req)
			token := vars["yubikey"]
			if token == "" {
				log.Warningf("Missing yubikey token")
				context.Fail(ctx, req, resp, "Invalid yubikey token", http.StatusUnauthorized)
				return
			}
			if len(token) != 44 {
				log.Warningf("Invalid yubikey token : %s", token)
				context.Fail(ctx, req, resp, "Invalid yubikey token", http.StatusUnauthorized)
				return
			}
			if token[:12] != upload.Yubikey {
				log.Warningf("Invalid yubikey device : %s", token)
				context.Fail(ctx, req, resp, "Invalid yubikey token", http.StatusUnauthorized)
				return
			}

			_, isValid, err := config.GetYubiAuth().Verify(token)
			if err != nil {
				log.Warningf("Failed to validate yubikey token : %s", err)
				context.Fail(ctx, req, resp, "Invalid yubikey token", http.StatusInternalServerError)
				return
			}
			if !isValid {
				log.Warningf("Invalid yubikey token : %s", token)
				context.Fail(ctx, req, resp, "Invalid yubikey token", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(resp, req)
	})
}
