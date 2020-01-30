package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	api_oauth2 "google.golang.org/api/oauth2/v2"
)

var oauth2TestEndpoint = oauth2.Endpoint{
	AuthURL:  "http://127.0.0.1:" + strconv.Itoa(common.APIMockServerDefaultPort),
	TokenURL: "http://127.0.0.1:" + strconv.Itoa(common.APIMockServerDefaultPort) + "/token",
}

func TestGoogleLogin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_app_id"
	ctx.GetConfig().GoogleAPISecret = "google_app_secret"

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	origin := "https://plik.root.gg"
	req.Header.Set("referer", origin)

	rr := httptest.NewRecorder()
	GoogleLogin(ctx, rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, 0, len(respBody), "invalid empty response body")

	URL, err := url.Parse(string(respBody))
	require.NoError(t, err, "unable to parse google auth url")

	state, err := jwt.Parse(URL.Query().Get("state"), func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("Unexpected siging method : %v", token.Header["alg"])
		}

		// Verify expiration data
		if expire, ok := token.Claims.(jwt.MapClaims)["expire"]; ok {
			if _, ok = expire.(float64); ok {
				if time.Now().Unix() > (int64)(expire.(float64)) {
					t.Fatal("state expired")
				}
			} else {
				t.Fatal("invalid state expiration date")
			}
		} else {
			t.Fatal("Missing state expiration date")
		}

		return []byte(ctx.GetConfig().GoogleAPISecret), nil
	})
	require.NoError(t, err, "invalid oauth2 state")

	require.Equal(t, origin, state.Claims.(jwt.MapClaims)["origin"].(string), "invalid state origin")

}

func TestGoogleLoginAuthDisabled(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = false
	ctx.GetConfig().GoogleAuthentication = false

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleLogin(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Authentication is disabled")
}

func TestGoogleLoginGoogleAuthDisabled(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = false

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("referer", "http://plik.root.gg")

	rr := httptest.NewRecorder()
	GoogleLogin(ctx, rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Missing google API credentials")
}

func TestGoogleLoginMissingReferer(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleLogin(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Missing referer header")
}

func TestGoogleCallback(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = "origin"
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	ctx.Set(googeleEndpointContextKey, oauth2TestEndpoint)

	oauthToken := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
	}{
		AccessToken:  "access_token",
		TokenType:    "token_type",
		RefreshToken: "refresh_token",
		ExpiresIn:    int32(time.Now().Add(5 * time.Minute).Unix()),
	}

	googleUser := api_oauth2.Userinfoplus{
		Id:    "plik",
		Email: "plik@root.gg",
		Name:  "plik.root.gg",
	}

	user := common.NewUser()
	user.ID = "google:plik"
	user.Login = googleUser.Email
	user.Name = googleUser.Email
	user.Email = googleUser.Email
	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add test user")

	handler := func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/token" {
			responseBody, err := json.Marshal(oauthToken)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		if req.URL.Path == "/oauth2/v2/userinfo" {
			responseBody, err := json.Marshal(googleUser)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		resp.WriteHeader(http.StatusInternalServerError)
	}

	shutdown, err := common.StartAPIMockServer(http.HandlerFunc(handler))
	defer shutdown()
	require.NoError(t, err, "unable to start ovh api mock server")

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, 301, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, 0, len(respBody), "invalid empty response body")

	var sessionCookie string
	var xsrfCookie string
	a := rr.Result().Cookies()
	require.NotEmpty(t, a)
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "plik-session" {
			sessionCookie = cookie.Value
		}
		if cookie.Name == "plik-xsrf" {
			xsrfCookie = cookie.Value
		}
	}

	require.NotEqual(t, "", sessionCookie, "missing plik session cookie")
	require.NotEqual(t, "", xsrfCookie, "missing plik xsrf cookie")
}

func TestGoogleCallbackAuthDisabled(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = false

	req, err := http.NewRequest("GET", "/auth/google/callback", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("referer", "http://plik.root.gg")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Authentication is disabled")
}

func TestGoogleCallbackMissingGoogleAuthParams(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("referer", "http://plik.root.gg")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Missing google API credentials")
}

func TestGoogleCallbackMissingCode(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	req, err := http.NewRequest("GET", "/auth/google/login", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Missing oauth2 authorization code")
}

func TestGoogleCallbackMissingState(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	req, err := http.NewRequest("GET", "/auth/google/login?code=code", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Missing oauth2 state")
}

func TestGoogleCallbackInvalidState(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state=state", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackExpiredState(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(-time.Minute * 5).Unix()

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackInvalidStateExpirationDate(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["expire"] = "invalid expiration date"

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackMissingStateExpirationDate(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackMissingOrigin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackInvalidOrigin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = -1
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Invalid oauth2 state")
}

func TestGoogleCallbackNoApi(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = "origin"
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	ctx.Set(googeleEndpointContextKey, oauth2TestEndpoint)

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to get user info from google API")
}

func TestGoogleCallbackCreateUser(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = "origin"
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	ctx.Set(googeleEndpointContextKey, oauth2TestEndpoint)

	oauthToken := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
	}{
		AccessToken:  "access_token",
		TokenType:    "token_type",
		RefreshToken: "refresh_token",
		ExpiresIn:    int32(time.Now().Add(5 * time.Minute).Unix()),
	}

	googleUser := api_oauth2.Userinfoplus{
		Id:    "plik",
		Email: "plik@root.gg",
		Name:  "plik.root.gg",
	}

	handler := func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/token" {
			responseBody, err := json.Marshal(oauthToken)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		if req.URL.Path == "/oauth2/v2/userinfo" {
			responseBody, err := json.Marshal(googleUser)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		resp.WriteHeader(http.StatusInternalServerError)
	}

	shutdown, err := common.StartAPIMockServer(http.HandlerFunc(handler))
	defer shutdown()
	require.NoError(t, err, "unable to start ovh api mock server")

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, 301, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, 0, len(respBody), "invalid empty response body")

	var sessionCookie string
	var xsrfCookie string
	a := rr.Result().Cookies()
	require.NotEmpty(t, a)
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "plik-session" {
			sessionCookie = cookie.Value
		}
		if cookie.Name == "plik-xsrf" {
			xsrfCookie = cookie.Value
		}
	}

	require.NotEqual(t, "", sessionCookie, "missing plik session cookie")
	require.NotEqual(t, "", xsrfCookie, "missing plik xsrf cookie")

	user, err := ctx.GetMetadataBackend().GetUser("google:plik")
	require.NotNil(t, user, "missing user")
	require.Equal(t, googleUser.Email, user.Email, "invalid user email")
	require.Equal(t, googleUser.Name, user.Name, "invalid user name")
}
func TestGoogleCallbackCreateUserNotWhitelisted(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetWhitelisted(ctx, false)

	ctx.GetConfig().Authentication = true
	ctx.GetConfig().GoogleAuthentication = true
	ctx.GetConfig().GoogleAPIClientID = "google_api_client_id"
	ctx.GetConfig().GoogleAPISecret = "google_api_secret"

	/* Generate state */
	state := jwt.New(jwt.SigningMethodHS256)
	state.Claims.(jwt.MapClaims)["origin"] = "origin"
	state.Claims.(jwt.MapClaims)["expire"] = time.Now().Add(time.Minute * 5).Unix()

	ctx.Set(googeleEndpointContextKey, oauth2TestEndpoint)

	oauthToken := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
	}{
		AccessToken:  "access_token",
		TokenType:    "token_type",
		RefreshToken: "refresh_token",
		ExpiresIn:    int32(time.Now().Add(5 * time.Minute).Unix()),
	}

	googleUser := api_oauth2.Userinfoplus{
		Id:    "plik",
		Email: "plik@root.gg",
		Name:  "plik.root.gg",
	}

	handler := func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/token" {
			responseBody, err := json.Marshal(oauthToken)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		if req.URL.Path == "/oauth2/v2/userinfo" {
			responseBody, err := json.Marshal(googleUser)
			require.NoError(t, err, "unable to marshal oauth token")
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseBody)
			return
		}
		resp.WriteHeader(http.StatusInternalServerError)
	}

	shutdown, err := common.StartAPIMockServer(http.HandlerFunc(handler))
	defer shutdown()
	require.NoError(t, err, "unable to start ovh api mock server")

	/* Sign state */
	b64state, err := state.SignedString([]byte(ctx.GetConfig().GoogleAPISecret))
	require.NoError(t, err, "unable to sign state")

	req, err := http.NewRequest("GET", "/auth/google/login?code=code&state="+url.QueryEscape(b64state), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GoogleCallback(ctx, rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "Unable to create user from untrusted source IP address")
}
