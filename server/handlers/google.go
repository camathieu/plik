package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	api_oauth2 "google.golang.org/api/oauth2/v2"
)

var googeleEndpointContextKey = "google_endpoint"

// GoogleLogin return google api user consent URL.
func GoogleLogin(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	if !config.Authentication {
		log.Warning("Authentication is disabled")
		context.Fail(ctx, req, resp, "Authentication is disabled", http.StatusBadRequest)
		return
	}

	if !config.GoogleAuthentication {
		log.Warning("Missing google api credentials")
		context.Fail(ctx, req, resp, "Missing google API credentials", http.StatusInternalServerError)
		return
	}

	origin := req.Header.Get("referer")
	if origin == "" {
		log.Warning("Missing referer header")
		context.Fail(ctx, req, resp, "Missing referer header", http.StatusBadRequest)
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
		log.Warningf("Unable to sign state : %s", err)
		context.Fail(ctx, req, resp, "Unable to sign state", http.StatusInternalServerError)
		return
	}

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(b64state)

	resp.Write([]byte(url))
}

// GoogleCallback authenticate google user.
func GoogleCallback(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	if !config.Authentication {
		log.Warning("Authentication is disabled")
		context.Fail(ctx, req, resp, "Authentication is disabled", http.StatusBadRequest)
		return
	}

	if config.GoogleAPIClientID == "" || config.GoogleAPISecret == "" {
		log.Warning("Missing google api credentials")
		context.Fail(ctx, req, resp, "Missing google API credentials", http.StatusInternalServerError)
		return
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		log.Warning("Missing oauth2 authorization code")
		context.Fail(ctx, req, resp, "Missing oauth2 authorization code", http.StatusBadRequest)
		return
	}

	b64state := req.URL.Query().Get("state")
	if b64state == "" {
		log.Warning("Missing oauth2 state")
		context.Fail(ctx, req, resp, "Missing oauth2 state", http.StatusBadRequest)
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
		log.Warningf("Invalid oauth2 state : %s", err)
		context.Fail(ctx, req, resp, "Invalid oauth2 state", http.StatusBadRequest)
		return
	}

	if _, ok := state.Claims.(jwt.MapClaims)["origin"]; !ok {
		log.Warning("Invalid oauth2 state : missing origin")
		context.Fail(ctx, req, resp, "Invalid oauth2 state", http.StatusBadRequest)
		return
	}

	if _, ok := state.Claims.(jwt.MapClaims)["origin"].(string); !ok {
		log.Warning("Invalid oauth2 state : invalid origin")
		context.Fail(ctx, req, resp, "Invalid oauth2 state", http.StatusBadRequest)
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

	if customEndpoint, ok := ctx.Get(googeleEndpointContextKey); ok {
		conf.Endpoint = customEndpoint.(oauth2.Endpoint)
	}

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Warningf("Unable to create google API token : %s", err)
		context.Fail(ctx, req, resp, "Unable to get user info from google API", http.StatusInternalServerError)
		return
	}

	client, err := api_oauth2.New(conf.Client(oauth2.NoContext, token))
	if err != nil {
		log.Warningf("Unable to create google API client : %s", err)
		context.Fail(ctx, req, resp, "Unable to get user info from google API", http.StatusInternalServerError)
		return
	}

	if customEndpoint, ok := ctx.Get(googeleEndpointContextKey); ok {
		client.BasePath = customEndpoint.(oauth2.Endpoint).AuthURL
	}

	userInfo, err := client.Userinfo.Get().Do()
	if err != nil {
		log.Warningf("Unable to get userinfo from google API : %s", err)
		context.Fail(ctx, req, resp, "Unable to get user info from google API", http.StatusInternalServerError)
		return
	}
	userID := "google:" + userInfo.Id

	// Get user from metadata backend
	user, err := context.GetMetadataBackend(ctx).GetUser(userID)
	if err != nil {
		log.Warningf("Unable to get user : %s", err)
		context.Fail(ctx, req, resp, "Unable to get user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		if context.IsWhitelisted(ctx) {
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
				log.Warningf("Unacceptable user domain : %s", components[1])
				context.Fail(ctx, req, resp, fmt.Sprintf("Authentication error : Unauthorized domain %s", components[1]), http.StatusForbidden)
				return
			}

			// Save user to metadata backend
			err = context.GetMetadataBackend(ctx).CreateUser(user)
			if err != nil {
				log.Warningf("Unable to save user to metadata backend : %s", err)
				context.Fail(ctx, req, resp, "Authentication error", http.StatusForbidden)
				return
			}
		} else {
			log.Warning("Unable to create user from untrusted source IP address")
			context.Fail(ctx, req, resp, "Unable to create user from untrusted source IP address", http.StatusForbidden)
			return
		}
	}

	// Set Plik session cookie and xsrf cookie
	sessionCookie, xsrfCookie, err := common.GenAuthCookies(user, context.GetConfig(ctx))
	if err != nil {
		log.Warningf("Unable to generate session cookies : %s", err)
		context.Fail(ctx, req, resp, "Authentication error", http.StatusForbidden)
	}
	http.SetCookie(resp, sessionCookie)
	http.SetCookie(resp, xsrfCookie)

	http.Redirect(resp, req, config.Path+"/#/login", http.StatusMovedPermanently)
}
