package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// CreateToken create a new token
func CreateToken(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	// Read request body
	defer func() { _ = req.Body.Close() }()

	req.Body = http.MaxBytesReader(resp, req.Body, 1048576)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ctx.BadRequest(fmt.Sprintf("unable to read request body : %s", err))
		return
	}

	// Create token
	token := common.NewToken()

	// Deserialize json body
	if len(body) > 0 {
		err = json.Unmarshal(body, token)
		if err != nil {
			ctx.BadRequest(fmt.Sprintf("unable to deserialize request body : %s", err))
			return
		}
	}

	// Generate token uuid and set creation date
	token.Initialize()

	// Add token to user
	tx := func(u *common.User) error {
		if u == nil {
			return common.NewHTTPError("user does not exist anymore", http.StatusNotFound)
		}
		u.Tokens = append(u.Tokens, token)
		return nil
	}

	// Save user
	user, err = ctx.GetMetadataBackend().UpdateUser(user, tx)
	if err != nil {
		handleTxError(ctx, "unable to update user metadata", err)
		return
	}

	// Print token in the json response.
	var bytes []byte
	if bytes, err = utils.ToJson(token); err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(bytes)
}

// RevokeToken remove a token
func RevokeToken(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	// Get token to remove from URL params
	vars := mux.Vars(req)
	tokenStr, ok := vars["token"]
	if !ok || tokenStr == "" {
		ctx.MissingParameter("token")
		return
	}

	// Remove token from user
	tx := func(u *common.User) error {
		if u == nil {
			return common.NewHTTPError("user does not exist anymore", http.StatusNotFound)
		}

		// Get token index
		index := -1
		for i, t := range u.Tokens {
			if t.Token == tokenStr {
				index = i
				break
			}
		}
		if index < 0 {
			return common.NewHTTPError(fmt.Sprintf("unable to get token %s from user %s", tokenStr, user.ID), http.StatusNotFound)
		}

		// Delete token
		u.Tokens = append(u.Tokens[:index], u.Tokens[index+1:]...)

		return nil
	}

	// Save user
	user, err := ctx.GetMetadataBackend().UpdateUser(user, tx)
	if err != nil {
		handleTxError(ctx, "unable to update user metadata", err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}
