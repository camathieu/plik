package middleware

import (
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// Authenticate verify that a request has either a whitelisted url or a valid auth token
func Authenticate(allowToken bool) juliet.ContextMiddleware {
	return func(ctx *juliet.Context, next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			log := context.GetLogger(ctx)
			config := context.GetConfig(ctx)

			if config.Authentication {
				if allowToken {
					// Get user from token header
					tokenHeader := req.Header.Get("X-PlikToken")
					if tokenHeader != "" {
						user, err := context.GetMetadataBackend(ctx).GetUserFromToken(tokenHeader)
						if err != nil {
							log.Warningf("Unable to get user from token %s : %s", tokenHeader, err)
							context.Fail(ctx, req, resp, "Unable to get user", http.StatusInternalServerError)
							return
						}
						if user == nil {
							log.Warningf("Unable to get user from token %s", tokenHeader)
							context.Fail(ctx, req, resp, "Invalid token", http.StatusForbidden)
							return
						}

						// Get token from user
						var token *common.Token
						for _, t := range user.Tokens {
							if t.Token == tokenHeader {
								token = t
								break
							}
						}
						if token == nil {
							// THIS SHOULD NEVER HAPPEN
							log.Warningf("Unable to get token %s from user %s", tokenHeader, user.ID)
							context.Fail(ctx, req, resp, "Invalid token", http.StatusInternalServerError)
							return
						}

						// Save user and token in the request context
						context.SetUser(ctx, user)
						context.SetToken(ctx, token)

						next.ServeHTTP(resp, req)
						return
					}
				}

				sessionCookie, err := req.Cookie("plik-session")
				if err == nil && sessionCookie != nil {
					// Parse session cookie
					uid, xsrf, err := common.ParseSessionCookie(sessionCookie.Value, config)
					if err != nil {
						log.Warningf("Invalid session : %s", err)
						common.Logout(resp)
						context.Fail(ctx, req, resp, "Invalid session", http.StatusForbidden)
						return
					}

					// Verify XSRF token
					if req.Method != "GET" && req.Method != "HEAD" {
						xsrfHeader := req.Header.Get("X-XSRFToken")
						if xsrfHeader == "" {
							log.Warning("Missing xsrf header")
							common.Logout(resp)
							context.Fail(ctx, req, resp, "Missing xsrf header", http.StatusForbidden)
							return
						}
						if xsrf != xsrfHeader {
							log.Warning("Invalid xsrf header")
							common.Logout(resp)
							context.Fail(ctx, req, resp, "Invalid xsrf header", http.StatusForbidden)
							return
						}
					}

					// Get user from session
					user, err := context.GetMetadataBackend(ctx).GetUser(uid)
					if err != nil {
						log.Warningf("Unable to get user from session : %s", err)
						common.Logout(resp)
						context.Fail(ctx, req, resp, "Unable to get user", http.StatusInternalServerError)
						return
					}
					if user == nil {
						log.Warningf("Invalid session : user does not exists")
						common.Logout(resp)
						context.Fail(ctx, req, resp, "Invalid session : User does not exists", http.StatusForbidden)
						return
					}

					// Save user in the request context
					context.SetUser(ctx, user)

					// Authenticate admin users
					if config.IsUserAdmin(user) {
						context.SetAdmin(ctx, true)
					}
				}
			}

			next.ServeHTTP(resp, req)
		})
	}
}
