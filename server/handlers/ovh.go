/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015> Copyright holders list can be found in AUTHORS file
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package handlers

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

type ovhError struct {
	ErrorCode string `json:"errorCode"`
	HTTPCode  string `json:"httpCode"`
	Message   string `json:"message"`
}

type ovhUserConsentResponse struct {
	ValidationURL string `json:"validationUrl"`
	ConsumerKey   string `json:"consumerKey"`
}

type ovhUserResponse struct {
	Nichandle string `json:"nichandle"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"name"`
}

func decodeOVHResponse(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body : %s", err)
	}

	if resp.StatusCode > 399 {
		// Decode OVH error information from response
		if body != nil && len(body) > 0 {
			var ovhErr ovhError
			err := json.Unmarshal(body, &ovhErr)
			if err == nil {
				return nil, fmt.Errorf("%s : %s", resp.Status, ovhErr.Message)
			}
			return nil, fmt.Errorf("%s : %s : %s", resp.Status, "Unable to unserialize ovh error", string(body))
		}
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return body, nil
}

// OvhLogin return ovh api user consent URL.
func OvhLogin(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	if !config.Authentication {
		log.Warning("Authentication is disabled")
		context.Fail(ctx, req, resp, "Authentication is disabled", http.StatusBadRequest)
		return
	}

	if !config.OvhAuthentication {
		log.Warning("Missing ovh api credentials")
		context.Fail(ctx, req, resp, "Missing OVH API credentials", http.StatusInternalServerError)
		return
	}

	origin := req.Header.Get("referer")
	if origin == "" {
		log.Warning("Missing referer header")
		context.Fail(ctx, req, resp, "Missing referer header", http.StatusBadRequest)
		return
	}

	// Prepare request
	redirectionURL := origin + "auth/ovh/callback"
	ovhReqBody := "{\"accessRules\":[{\"method\":\"GET\",\"path\":\"/me\"}], \"redirection\":\"" + redirectionURL + "\"}"

	url := fmt.Sprintf("%s/auth/credential", config.OvhAPIEndpoint)

	ovhReq, err := http.NewRequest("POST", url, strings.NewReader(ovhReqBody))
	ovhReq.Header.Add("X-Ovh-Application", config.OvhAPIKey)
	ovhReq.Header.Add("Content-type", "application/json")

	// Do request
	client := &http.Client{}
	ovhResp, err := client.Do(ovhReq)
	if err != nil {
		log.Warningf("Error with ovh API %s : %s", url, err)
		context.Fail(ctx, req, resp, "Error with OVH API ", http.StatusInternalServerError)
		return
	}
	defer ovhResp.Body.Close()
	ovhRespBody, err := decodeOVHResponse(ovhResp)
	if err != nil {
		log.Warningf("Error with ovh API %s : %s", url, err)
		context.Fail(ctx, req, resp, fmt.Sprintf("Error with OVH API : %s", err), http.StatusInternalServerError)
		return
	}

	var userConsentResponse ovhUserConsentResponse
	err = json.Unmarshal(ovhRespBody, &userConsentResponse)
	if err != nil {
		log.Warningf("Unable to unserialize OVH API response : %s", err)
		context.Fail(ctx, req, resp, "Unable to unserialize OVH API response", http.StatusInternalServerError)
		return
	}

	// Generate session jwt
	session := jwt.New(jwt.SigningMethodHS256)
	session.Claims.(jwt.MapClaims)["ovh-consumer-key"] = userConsentResponse.ConsumerKey
	session.Claims.(jwt.MapClaims)["ovh-api-endpoint"] = config.OvhAPIEndpoint

	sessionString, err := session.SignedString([]byte(config.OvhAPISecret))
	if err != nil {
		log.Warningf("Unable to sign OVH session cookie : %s", err)
		context.Fail(ctx, req, resp, "Unable to sign OVH session cookie", http.StatusInternalServerError)
		return
	}

	// Store temporary session jwt in secure cookie
	ovhAuthCookie := &http.Cookie{}
	ovhAuthCookie.HttpOnly = true
	ovhAuthCookie.Secure = true
	ovhAuthCookie.Name = "plik-ovh-session"
	ovhAuthCookie.Value = sessionString
	ovhAuthCookie.MaxAge = int(time.Now().Add(5 * time.Minute).Unix())
	ovhAuthCookie.Path = "/"
	http.SetCookie(resp, ovhAuthCookie)

	resp.Write([]byte(userConsentResponse.ValidationURL))
}

// Remove temporary session cookie
func cleanOvhAuthSessionCookie(resp http.ResponseWriter) {
	ovhAuthCookie := &http.Cookie{}
	ovhAuthCookie.HttpOnly = true
	ovhAuthCookie.Secure = true
	ovhAuthCookie.Name = "plik-ovh-session"
	ovhAuthCookie.Value = ""
	ovhAuthCookie.MaxAge = -1
	ovhAuthCookie.Path = "/"
	http.SetCookie(resp, ovhAuthCookie)
}

// OvhCallback authenticate ovh user.
func OvhCallback(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	// Remove temporary ovh auth session cookie
	cleanOvhAuthSessionCookie(resp)

	if !config.Authentication {
		log.Warning("Authentication is disabled")
		context.Fail(ctx, req, resp, "Authentication is disabled", http.StatusBadRequest)
		return
	}

	if config.OvhAPIKey == "" || config.OvhAPISecret == "" || config.OvhAPIEndpoint == "" {
		log.Warning("Missing ovh api credentials")
		context.Fail(ctx, req, resp, "Missing OVH API credentials", http.StatusInternalServerError)
		return
	}

	// Get state from secure cookie
	ovhSessionCookie, err := req.Cookie("plik-ovh-session")
	if err != nil || ovhSessionCookie == nil {
		log.Warning("Missing OVH session cookie")
		context.Fail(ctx, req, resp, "Missing OVH session cookie", http.StatusBadRequest)
		return
	}

	// Parse session cookie
	ovhAuthCookie, err := jwt.Parse(ovhSessionCookie.Value, func(t *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected siging method : %v", t.Header["alg"])
		}

		return []byte(config.OvhAPISecret), nil
	})
	if err != nil {
		log.Warningf("Invalid OVH session cookie : %s", err)
		context.Fail(ctx, req, resp, "Invalid OVH session cookie", http.StatusBadRequest)
		return
	}

	// Get OVH consumer key from session
	ovhConsumerKey, ok := ovhAuthCookie.Claims.(jwt.MapClaims)["ovh-consumer-key"]
	if !ok {
		log.Warning("Invalid OVH session cookie : missing ovh-consumer-key")
		context.Fail(ctx, req, resp, "Invalid OVH session cookie : missing ovh-consumer-key", http.StatusBadRequest)
		return
	}

	// Get OVH API endpoint
	endpoint, ok := ovhAuthCookie.Claims.(jwt.MapClaims)["ovh-api-endpoint"]
	if !ok {
		log.Warning("Invalid OVH session cookie : missing ovh-api-endpoint")
		context.Fail(ctx, req, resp, "Invalid OVH session cookie : missing ovh-api-endpoint", http.StatusBadRequest)
		return
	}

	// Prepare OVH API /me request
	url := endpoint.(string) + "/me"
	ovhReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Warningf("Unable to create new http GET request to %s : %s", url, err)
		context.Fail(ctx, req, resp, "Unable to create new http GET request to OVH API", http.StatusInternalServerError)
		return
	}

	timestamp := time.Now().Unix()
	ovhReq.Header.Add("X-Ovh-Application", config.OvhAPIKey)
	ovhReq.Header.Add("X-Ovh-Timestamp", fmt.Sprintf("%d", timestamp))
	ovhReq.Header.Add("X-Ovh-Consumer", ovhConsumerKey.(string))

	// Sign request
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s+%s+%s+%s+%s+%d",
		config.OvhAPISecret,
		ovhConsumerKey.(string),
		"GET",
		url,
		"",
		timestamp,
	)))
	ovhReq.Header.Add("X-Ovh-Signature", fmt.Sprintf("$1$%x", h.Sum(nil)))

	// Do request
	client := &http.Client{}
	ovhResp, err := client.Do(ovhReq)
	if err != nil {
		log.Warningf("Error with ovh API %s : %s", url, err)
		context.Fail(ctx, req, resp, "Error with ovh API", http.StatusInternalServerError)
		return
	}
	defer ovhResp.Body.Close()
	ovhRespBody, err := decodeOVHResponse(ovhResp)
	if err != nil {
		log.Warningf("Error with ovh API %s : %s", url, err)
		context.Fail(ctx, req, resp, fmt.Sprintf("Error with OVH API : %s", err), http.StatusInternalServerError)
		return
	}

	// Unserialize response
	var userInfo ovhUserResponse
	err = json.Unmarshal(ovhRespBody, &userInfo)
	if err != nil {
		log.Warningf("Unable to unserialize OVH API response : %s", err)
		context.Fail(ctx, req, resp, "Unable to unserialize OVH API response", http.StatusInternalServerError)
		return
	}

	userID := "ovh:" + userInfo.Nichandle

	// Get user from metadata backend
	user, err := context.GetMetadataBackend(ctx).GetUser(userID)
	if err != nil {
		log.Warningf("Unable to get user from metadata backend : %s", err)
		context.Fail(ctx, req, resp, "Unable to get user from metadata backend", http.StatusInternalServerError)
		return
	}

	if user == nil {
		if context.IsWhitelisted(ctx) {
			// Create new user
			user = common.NewUser()
			user.ID = userID
			user.Login = userInfo.Nichandle
			user.Name = userInfo.FirstName + " " + userInfo.LastName
			user.Email = userInfo.Email

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
