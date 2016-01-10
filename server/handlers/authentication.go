/* The MIT License (MIT)

Copyright (c) <2015>
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
THE SOFTWARE. */

package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/authentication/google"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/metadataBackend"
)

// GetVersionHandler return the build information.
func LoginHandler(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := common.GetLogger(ctx)
	var err error
	var redirectUrl string

	origin := req.URL.Query().Get("origin")
	_, err = url.Parse(origin)
	if err != nil {
		log.Warning("Missing request origin")
		common.Fail(ctx, req, resp, "Unable to get request origin", 403)
		return
	}

	vars := mux.Vars(req)
	switch vars["provider"] {
	case "google":
		redirectUrl, err = google.GetUserConsentUrl(ctx, origin)
		if err != nil {
			common.Fail(ctx, req, resp, "Unable to get google user consent url", 403)
			return
		}
	default:
		common.Fail(ctx, req, resp, "Invalid authentication provider", 400)
		return
	}

	log.Debugf("login url : %s", redirectUrl)
	resp.Write([]byte(redirectUrl))
}

// GetVersionHandler return the build information.
func CallbackHandler(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := common.GetLogger(ctx)
	var err error
	var user *common.User

	vars := mux.Vars(req)
	switch vars["provider"] {
	case "google":
		user, err = google.Callback(ctx, req.URL.Query().Get("state"), req.URL.Query().Get("code"))
		if err != nil {
			log.Warningf("Unable to verify google callback : %s", err)
			common.Fail(ctx, req, resp, "Unable to get user consent url", 403)
			return
		}
	default:
		common.Fail(ctx, req, resp, "Invalid authentication provider", 400)
		return
	}

	// Generate a new token
	token := user.NewToken()

	// Save user to metadata backend
	err = metadataBackend.GetMetaDataBackend().SaveUser(ctx, user)
	if err != nil {
		log.Warningf("Unable to save user to metadata backend : %s", err)
		common.Fail(ctx, req, resp, "Authentification error", 403)
		return
	}

	http.Redirect(resp, req, fmt.Sprintf("/#/login?token=%s", token.Token), 301)
}
