package middleware

import (
	"fmt"
	"net/http"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// Authenticate verify that a request has either a whitelisted url or a valid auth token
func Authenticate(allowToken bool) context.ContextMiddleware {
	return func(ctx *context.Context, next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			config := ctx.GetConfig()

			if config.Authentication {
				if allowToken {
					// Get user from token header
					tokenHeader := req.Header.Get("X-PlikToken")
					if tokenHeader != "" {
						user, err := ctx.GetMetadataBackend().GetUserFromToken(tokenHeader)
						if err != nil {
							ctx.InternalServerError(fmt.Errorf("unable to get user from token : %s", err))
							return
						}
						if user == nil {
							ctx.Forbidden("invalid token")
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
							ctx.InternalServerError(fmt.Errorf("missing token %s from user %s", tokenHeader, user.ID))
							return
						}

						// Save user and token in the request context
						ctx.SetUser(user)
						ctx.SetToken(token)

						next.ServeHTTP(resp, req)
						return
					}
				}

				sessionCookie, err := req.Cookie("plik-session")
				if err == nil && sessionCookie != nil {
					// Parse session cookie
					uid, xsrf, err := common.ParseSessionCookie(sessionCookie.Value, config)
					if err != nil {
						common.Logout(resp)
						ctx.Forbidden("invalid session")
						return
					}

					// Verify XSRF token
					if req.Method != "GET" && req.Method != "HEAD" {
						xsrfHeader := req.Header.Get("X-XSRFToken")
						if xsrfHeader == "" {
							common.Logout(resp)
							ctx.Forbidden("missing xsrf header")
							return
						}
						if xsrf != xsrfHeader {
							common.Logout(resp)
							ctx.Forbidden("invalid xsrf header")
							return
						}
					}

					// Get user from session
					user, err := ctx.GetMetadataBackend().GetUser(uid)
					if err != nil {
						common.Logout(resp)
						ctx.InternalServerError(fmt.Errorf("unable to get user from session : %s", err))
						return
					}
					if user == nil {
						common.Logout(resp)
						ctx.Forbidden("invalid session : user does not exists")
						return
					}

					// Save user in the request context
					ctx.SetUser(user)

					// Authenticate admin users
					if config.IsUserAdmin(user) {
						ctx.SetAdmin(true)
					}
				}
			}

			next.ServeHTTP(resp, req)
		})
	}
}
