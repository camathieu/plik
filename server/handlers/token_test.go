package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"testing"

	"github.com/gorilla/mux"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	user.ID = "user1"

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	token := common.NewToken()
	token.Comment = "token comment"

	reqBody, err := json.Marshal(token)
	require.NoError(t, err, "unable to marshal request body")

	req, err := http.NewRequest("POST", "/me/token", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	CreateToken(ctx, rr, req)

	// Check the status code is what we expect.
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	var tokenResult = &common.Token{}
	err = json.Unmarshal(respBody, tokenResult)
	require.NoError(t, err, "unable to unmarshal response body")

	require.NotEqual(t, "", tokenResult.Token, "missing token id")
	require.NotEqual(t, 0, tokenResult.CreationDate, "missing token creation date")
	require.Equal(t, token.Comment, tokenResult.Comment, "invalid token comment")
}

func TestCreateTokenWithForbiddenOptions(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	user.ID = "user1"

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	token := common.NewToken()
	token.Comment = "token comment"
	token.Token = "invalid"
	token.CreationDate = -1

	reqBody, err := json.Marshal(token)
	require.NoError(t, err, "unable to marshal request body")

	req, err := http.NewRequest("POST", "/me/token", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	CreateToken(ctx, rr, req)

	// Check the status code is what we expect.
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	var tokenResult = &common.Token{}
	err = json.Unmarshal(respBody, tokenResult)
	require.NoError(t, err, "unable to unmarshal response body")

	require.NotEqual(t, token.Token, tokenResult.Token, "invalid token id")
	require.NotEqual(t, token.CreationDate, tokenResult.CreationDate, "invalid token creation date")
	require.Equal(t, token.Comment, tokenResult.Comment, "invalid token comment")
}

func TestCreateTokenMissingUser(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	req, err := http.NewRequest("GET", "/me/token", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	CreateToken(ctx, rr, req)
	context.TestFail(t, rr, http.StatusUnauthorized, "Please login first")
}

func TestCreateTokenMetadataBackendError(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	user := common.NewUser()
	user.ID = "user1"

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	token := common.NewToken()
	token.Comment = "token comment"

	reqBody, err := json.Marshal(token)
	require.NoError(t, err, "unable to marshal request body")

	req, err := http.NewRequest("POST", "/me/token", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "unable to create new request")

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := ctx.NewRecorder(req)
	CreateToken(ctx, rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to update user metadata")
}

func TestRemoveToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	user.ID = "user1"

	token := common.NewToken()
	token.Comment = "token comment"
	token.Create()
	user.Tokens = append(user.Tokens, token)

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	req, err := http.NewRequest("DELETE", "/me/token/"+token.Token, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"token": token.Token,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	RevokeToken(ctx, rr, req)

	// Check the status code is what we expect.
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody), "invalid response body")

	user, err = ctx.GetMetadataBackend().GetUser(user.ID)
	require.NoError(t, err, "unable to get user")
	require.Equal(t, 0, len(user.Tokens), "invalid user token count")
}

func TestRemoveMissingToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	user.ID = "user1"

	token := common.NewToken()
	token.Comment = "token comment"
	token.Create()
	user.Tokens = append(user.Tokens, token)

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	req, err := http.NewRequest("DELETE", "/me/token/invalid_token_id", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"token": "invalid_token_id",
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	RevokeToken(ctx, rr, req)

	context.TestFail(t, rr, http.StatusNotFound, "unable to get token")
}

func TestRevokeTokenMissingUser(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	req, err := http.NewRequest("DELETE", "/me/token", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RevokeToken(ctx, rr, req)
	context.TestFail(t, rr, http.StatusUnauthorized, "Please login first")
}

func TestRevokeTokenMetadataBackendError(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	user := common.NewUser()
	user.ID = "user1"

	token := common.NewToken()
	token.Comment = "token comment"
	token.Create()
	user.Tokens = append(user.Tokens, token)

	err := addTestUser(ctx, user)
	require.NoError(t, err, "unable to add user")
	ctx.SetUser(user)

	req, err := http.NewRequest("DELETE", "/me/token", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"token": token.Token,
	}
	req = mux.SetURLVars(req, vars)

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := ctx.NewRecorder(req)
	RevokeToken(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to update upload metadata")
}
