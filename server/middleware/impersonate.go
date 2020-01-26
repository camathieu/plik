package middleware

import (
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
)

// Impersonate allow an administrator to pretend being another user
func Impersonate(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := context.GetLogger(ctx)

		// Get user to impersonate from header
		newUserID := req.Header.Get("X-Plik-Impersonate")
		if newUserID != "" {

			// Check authorization
			if !context.IsAdmin(ctx) {
				log.Warningf("Unable to impersonate user : unauthorized")
				context.Fail(ctx, req, resp, "You need administrator privileges", http.StatusForbidden)
				return
			}

			newUser, err := context.GetMetadataBackend(ctx).GetUser(newUserID)
			if err != nil {
				log.Warningf("Unable to get user to impersonate %s : %s", newUserID, err)
				context.Fail(ctx, req, resp, "Unable to get user to impersonate", http.StatusInternalServerError)
				return
			}

			if newUser == nil {
				log.Warningf("Unable to get user to impersonate : user does not exists")
				context.Fail(ctx, req, resp, "Unable to get user to impersonate : User does not exists", http.StatusForbidden)
				return
			}

			// Change user in the request context
			context.SetUser(ctx, newUser)
		}

		next.ServeHTTP(resp, req)
	})
}
