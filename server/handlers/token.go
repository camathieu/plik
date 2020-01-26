package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// CreateToken create a new token
func CreateToken(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get user from context
	user := context.GetUser(ctx)
	if user == nil {
		context.Fail(ctx, req, resp, "Missing user, Please login first", http.StatusUnauthorized)
		return
	}

	// Create token
	token := common.NewToken()

	// Read request body
	defer req.Body.Close()
	req.Body = http.MaxBytesReader(resp, req.Body, 1048576)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warningf("Unable to read request body : %s", err)
		context.Fail(ctx, req, resp, "Unable to read request body", http.StatusForbidden)
		return
	}

	// Deserialize json body
	if len(body) > 0 {
		err = json.Unmarshal(body, token)
		if err != nil {
			log.Warningf("Unable to deserialize json request body : %s", err)
			context.Fail(ctx, req, resp, "Unable to deserialize json request body", http.StatusBadRequest)
			return
		}
	}

	// Initialize token
	token.Create()

	// Add token to user
	tx := func(u *common.User) error {
		u.Tokens = append(u.Tokens, token)
		return nil
	}

	// Save user
	user, err = context.GetMetadataBackend(ctx).UpdateUser(user, tx)
	if err != nil {
		log.Warningf("Unable update user metadata : %s", err)
		context.Fail(ctx, req, resp, "Unable to update user metadata", http.StatusInternalServerError)
		return
	}

	// Print token in the json response.
	var json []byte
	if json, err = utils.ToJson(token); err != nil {
		log.Warningf("Unable to serialize json response : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
		return
	}
	resp.Write(json)
}

// RevokeToken remove a token
func RevokeToken(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get user from context
	user := context.GetUser(ctx)
	if user == nil {
		context.Fail(ctx, req, resp, "Missing user, Please login first", http.StatusUnauthorized)
		return
	}

	// Get token to remove from URL params
	vars := mux.Vars(req)
	tokenStr, ok := vars["token"]
	if !ok || tokenStr == "" {
		context.Fail(ctx, req, resp, "Missing token", http.StatusBadRequest)
	}

	// Remove token from user
	tx := func(u *common.User) error {
		// Get token index
		index := -1
		for i, t := range u.Tokens {
			if t.Token == tokenStr {
				index = i
				break
			}
		}
		if index < 0 {
			return common.NewTxError(fmt.Sprintf("unable to get token %s from user %s", tokenStr, user.ID), http.StatusNotFound)
		}

		// Delete token
		u.Tokens = append(u.Tokens[:index], u.Tokens[index+1:]...)

		return nil
	}

	// Save user
	user, err := context.GetMetadataBackend(ctx).UpdateUser(user, tx)
	if err != nil {
		if txError, ok := err.(common.TxError); ok {
			context.Fail(ctx, req, resp, txError.Error(), txError.GetStatusCode())
			return
		} else {
			log.Warningf("Unable to update upload metadata : %s", err)
			context.Fail(ctx, req, resp, "Unable to update upload metadata", http.StatusInternalServerError)
			return
		}
	}

	_, _ = resp.Write([]byte("ok"))
}
