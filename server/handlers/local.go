package handlers

import (
	"encoding/json"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"io/ioutil"
	"net/http"
)

type LoginParams struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

func LocalLogin(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	if !config.Authentication {
		ctx.BadRequest("authentication is disabled")
		return
	}

	// Read request body
	defer func() { _ = req.Body.Close() }()
	req.Body = http.MaxBytesReader(resp, req.Body, 1048576)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ctx.BadRequest("unable to read request body : %s", err)
		return
	}

	loginParams := &LoginParams{}
	err = json.Unmarshal(body, loginParams)
	if err != nil {
		ctx.BadRequest("unable to deserialize request body : %s", err)
		return
	}

	if loginParams.Login == "" {
		ctx.MissingParameter("login")
		return
	}

	if loginParams.Password == "" {
		ctx.MissingParameter("password")
		return
	}

	// Get user from metadata backend
	user, err := ctx.GetMetadataBackend().GetUser(common.GetUserId(common.ProviderLocal, loginParams.Login))
	if err != nil {
		ctx.InternalServerError("unable to get user from metadata backend", err)
		return
	}

	if user == nil {
		ctx.Forbidden("invalid credentials")
		return
	}

	if !common.CheckPasswordHash(loginParams.Password, user.Password) {
		ctx.Forbidden("invalid credentials")
		return
	}

	// Set Plik session cookie and xsrf cookie
	sessionCookie, xsrfCookie, err := common.GenAuthCookies(user, ctx.GetConfig())
	if err != nil {
		ctx.InternalServerError("unable to generate session cookies", err)
	}
	http.SetCookie(resp, sessionCookie)
	http.SetCookie(resp, xsrfCookie)

	_,_ = resp.Write([]byte("ok"))
}
