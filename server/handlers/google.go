package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	api_oauth2 "google.golang.org/api/oauth2/v2"
)

var googeleEndpointContextKey = "google_endpoint"

// GoogleLogin return google api user consent URL.
func GoogleLogin(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	if !config.Authentication {
		ctx.Forbidden("authentication is disabled")
		return
	}

	if !config.GoogleAuthentication {
		ctx.Forbidden("Google authentication is disabled")
		return
	}

	origin := req.Header.Get("referer")
	if origin == "" {
		ctx.MissingParameter("referer")
		return
	}

	conf := &oauth2.Config{
		ClientID:     config.GoogleAPIClientID,
		ClientSecret: config.GoogleAPISecret,
		RedirectURL:  origin + "auth/google/callback",
		Scopes: []string{
			api_oauth2.UserinfoEmailScope,
			api_oauth2.UserinfoProfileScope,
		},
		Endpoint: google.Endpoint,
	}

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = origin
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	/* Sign state */
	b64state, err := state.SignedString([]byte(config.GoogleAPISecret))
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to sign state : %s", err))
		return
	}

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(b64state)

	_, _ = resp.Write([]byte(url))
}

// GoogleCallback authenticate google user.
func GoogleCallback(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	if !config.Authentication {
		ctx.Forbidden("authentication is disabled")
		return
	}

	if !config.GoogleAuthentication {
		ctx.Forbidden("Google authentication is disabled")
		return
	}

	if config.GoogleAPIClientID == "" || config.GoogleAPISecret == "" {
		ctx.InternalServerError(fmt.Errorf("missing Google API credentials"))
		return
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		ctx.MissingParameter("oauth2 authorization code")
		return
	}

	b64state := req.URL.Query().Get("state")
	if b64state == "" {
		ctx.MissingParameter("oauth2 authorization state")
		return
	}

	/* Parse state */
	state, err := jwt.Parse(b64state, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected siging method : %v", token.Header["alg"])
		}

		// Verify expiration data
		if expire, ok := token.Claims.(jwt.MapClaims)["expire"]; ok {
			if _, ok = expire.(float64); ok {
				if time.Now().Unix() > (int64)(expire.(float64)) {
					return nil, fmt.Errorf("State has expired")
				}
			} else {
				return nil, fmt.Errorf("Invalid expiration date")
			}
		} else {
			return nil, fmt.Errorf("Missing expiration date")
		}

		return []byte(config.GoogleAPISecret), nil
	})
	if err != nil {
		ctx.InvalidParameter(fmt.Sprintf("oauth2 state : %s", err))
		return
	}

	if _, ok := state.Claims.(jwt.MapClaims)["origin"]; !ok {
		ctx.InvalidParameter("oauth2 state : missing origin")
		return
	}

	if _, ok := state.Claims.(jwt.MapClaims)["origin"].(string); !ok {
		ctx.InvalidParameter("oauth2 state : invalid origin")
		return
	}

	origin := state.Claims.(jwt.MapClaims)["origin"].(string)

	conf := &oauth2.Config{
		ClientID:     config.GoogleAPIClientID,
		ClientSecret: config.GoogleAPISecret,
		RedirectURL:  origin + "auth/google/callback",
		Scopes: []string{
			api_oauth2.UserinfoEmailScope,
			api_oauth2.UserinfoProfileScope,
		},
		Endpoint: google.Endpoint,
	}

	// For testing purpose
	if customEndpoint := req.Context().Value(googeleEndpointContextKey); customEndpoint != nil {
		conf.Endpoint = customEndpoint.(oauth2.Endpoint)
	}

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to get user info from Google API : %s", err))
		return
	}

	client, err := api_oauth2.New(conf.Client(oauth2.NoContext, token))
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to get user info from Google API : %s", err))
		return
	}

	// For testing purpose
	if customEndpoint := req.Context().Value(googeleEndpointContextKey); customEndpoint != nil {
		client.BasePath = customEndpoint.(oauth2.Endpoint).AuthURL
	}

	userInfo, err := client.Userinfo.Get().Do()
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to get user info from Google API : %s", err))
		return
	}
	userID := "google:" + userInfo.Id

	// Get user from metadata backend
	user, err := ctx.GetMetadataBackend().GetUser(userID)
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to get user from metadata backend : %s", err))
		return
	}

	if user == nil {
		if ctx.IsWhitelisted() {
			// Create new user
			user = common.NewUser()
			user.ID = userID
			user.Login = userInfo.Email
			user.Name = userInfo.Name
			user.Email = userInfo.Email
			components := strings.Split(user.Email, "@")

			// Accepted user domain checking
			goodDomain := false
			if len(config.GoogleValidDomains) > 0 {
				for _, validDomain := range config.GoogleValidDomains {
					if strings.Compare(components[1], validDomain) == 0 {
						goodDomain = true
					}
				}
			} else {
				goodDomain = true
			}

			if !goodDomain {
				// User not from accepted google domains list
				ctx.Forbidden("unauthorized domain name")
				return
			}

			// Save user to metadata backend
			err = ctx.GetMetadataBackend().CreateUser(user)
			if err != nil {
				ctx.InternalServerError(fmt.Errorf("unable to create user in metadata backend : %s", err))
				return
			}
		} else {
			ctx.Forbidden("unable to create user from untrusted source IP address")
			return
		}
	}

	// Set Plik session cookie and xsrf cookie
	sessionCookie, xsrfCookie, err := common.GenAuthCookies(user, ctx.GetConfig())
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to generate session cookies : %s", err))
	}
	http.SetCookie(resp, sessionCookie)
	http.SetCookie(resp, xsrfCookie)

	http.Redirect(resp, req, config.Path+"/#/login", http.StatusMovedPermanently)
}
