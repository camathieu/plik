package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

//
// User registration is a 3 steps process
//  - 1 : registration : Create the user ( this might require a valid invite code )
//  - 2 : confirmation : Send a confirmation code by email
//  - 3 : verification : The user input the confirmation code recieved
//

// LoginParams to be POSTed by clients to authenticate
type RegisterParams struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty" gorm:"unique"`
}

func Register(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	if !ctx.GetConfig().Authentication {
		ctx.BadRequest("authentication is disabled")
		return
	}

	if ctx.GetConfig().Registration == common.RegistrationClosed {
		ctx.BadRequest("user registration is disabled")
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

	// Deserialize json body
	params := &RegisterParams{}
	err = json.Unmarshal(body, params)
	if err != nil {
		ctx.BadRequest("unable to deserialize request body : %s", err)
		return
	}

	user := common.NewUser(common.ProviderLocal, params.Login)
	user.Login = params.Login
	user.Name = params.Name
	user.Email = params.Email
	user.Password = params.Password

	err = user.PrepareInsert(ctx.GetConfig())
	if err != nil {
		ctx.BadRequest("unable to create user : %s", err)
		return
	}

	// Check if user already exists
	// TODO Gorm does not currently provide a way to handle unique constraint violation as a specific error
	// See : https://github.com/go-gorm/gorm/pull/3512
	// So we'll do a quick check to return a nice error message but there is obviously a small race condition

	u, err := ctx.GetMetadataBackend().GetUser(user.ID)
	if err != nil {
		ctx.BadRequest("unable to get user : %s", err)
		return
	}
	if u != nil {
		ctx.BadRequest("user already exists")
		return
	}

	if !ctx.GetConfig().EmailVerification {
		user.Verified = true
	}

	err = ctx.GetMetadataBackend().CreateUser(user)
	if err != nil {
		ctx.InternalServerError("unable to create user : %s", err)
		return
	}

	// Set Plik session cookie and xsrf cookie
	sessionCookie, xsrfCookie, err := ctx.GetAuthenticator().GenAuthCookies(user)
	if err != nil {
		ctx.InternalServerError("unable to generate session cookies", err)
		return
	}
	http.SetCookie(resp, sessionCookie)
	http.SetCookie(resp, xsrfCookie)

	common.WriteJSONResponse(resp, user)
}

func Confirm(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	user.GenVerificationCode()

	err := ctx.GetMetadataBackend().UpdateUser(user)
	if err != nil {
		ctx.InternalServerError("unable to update user metadata : %s", err)
		return
	}

	url := fmt.Sprintf("%s/auth/local/verify/%s/%s", ctx.GetConfig().GetServerURL().String(), user.Login, user.VerificationCode)
	common.WriteStringResponse(resp, url)
}

func Verify(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user
	// Get the file id from the url params
	vars := mux.Vars(req)
	userID := vars["userID"]
	if userID == "" {
		ctx.MissingParameter("user ID")
		return
	}

	user, err := ctx.GetMetadataBackend().GetUser(common.GetUserID(common.ProviderLocal, userID))
	if err != nil {
		ctx.BadRequest("unable to get user : %s", err)
		return
	}

	if user == nil {
		ctx.BadRequest("user does not exists")
		return
	}

	if user.VerificationCode == "" {
		ctx.BadRequest("missing confirmation code, please send confirmation code first")
		return
	}

	// Get the file id from the url params
	code := vars["code"]
	if code == "" {
		ctx.MissingParameter("verification code")
		return
	}

	if user.VerificationCode != code {
		ctx.Unauthorized("invalid verification code")
		return
	}

	user.VerificationCode = ""
	user.Verified = true

	err = ctx.GetMetadataBackend().UpdateUser(user)
	if err != nil {
		ctx.InternalServerError("unable to update user metadata : %s", err)
		return
	}

	// Set Plik session cookie and xsrf cookie
	sessionCookie, xsrfCookie, err := ctx.GetAuthenticator().GenAuthCookies(user)
	if err != nil {
		ctx.InternalServerError("unable to generate session cookies", err)
		return
	}
	http.SetCookie(resp, sessionCookie)
	http.SetCookie(resp, xsrfCookie)

	http.Redirect(resp, req, ctx.GetConfig().Path+"/#/login", http.StatusMovedPermanently)
}
