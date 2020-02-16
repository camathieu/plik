package handlers

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
	"net/http"
)

// UserInfo return user information ( name / email / ... )
func UserInfo(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	// Get user tokens ( this should be a separate call )
	tokens, err := ctx.GetMetadataBackend().GetTokens(user)
	if err != nil {
		ctx.InternalServerError("unable to get user tokens", err)
		return
	}

	user.Tokens = tokens

	// Serialize user to JSON
	// Print token in the json response.
	json, err := utils.ToJson(user)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// UserTokens return user tokens
func UserTokens(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	// Get user tokens
	tokens, err := ctx.GetMetadataBackend().GetTokens(user)
	if err != nil {
		ctx.InternalServerError("unable to get user tokens", err)
		return
	}

	// Serialize user to JSON
	// Print token in the json response.
	json, err := utils.ToJson(tokens)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// DeleteAccount remove a user account
func DeleteAccount(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	err := ctx.GetMetadataBackend().DeleteUser(user.ID)
	if err != nil {
		ctx.InternalServerError("unable to delete user account", err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}

// GetUserUploads get user uploads
func GetUserUploads(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	user, token, err := getUserAndToken(ctx, req)
	if err != nil {
		handleHTTPError(ctx, err)
		return
	}

	pagingQuery := ctx.GetPagingQuery()
	if pagingQuery == nil {
		ctx.InternalServerError("missing paging query", nil)
		return
	}

	var userID, tokenStr string
	if user != nil {
		userID = user.ID
	}
	if token != nil {
		tokenStr = token.Token
	}

	// Get uploads
	uploads, cursor, err := ctx.GetMetadataBackend().GetUploads(userID, tokenStr, true, pagingQuery)
	if err != nil {
		ctx.InternalServerError("unable to get user uploads : %s", err)
		return
	}

	pr := common.NewPagingResponse(uploads, cursor)

	// Serialize user to JSON
	// Print token in the json response.
	json, err := utils.ToJson(pr)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// RemoveUserUploads delete all user uploads
func RemoveUserUploads(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	user, token, err := getUserAndToken(ctx, req)
	if err != nil {
		handleHTTPError(ctx, err)
		return
	}

	var userID, tokenStr string
	if user != nil {
		userID = user.ID
	}
	if token != nil {
		tokenStr = token.Token
	}

	deleted, err := ctx.GetMetadataBackend().DeleteUserUploads(userID, tokenStr)
	if err != nil {
		ctx.InternalServerError("unable to delete user uploads", err)
		return
	}

	_, _ = resp.Write(common.NewResult(fmt.Sprintf("%d uploads removed", deleted), nil).ToJSON())
}

// GetUserStatistics return the user statistics
func GetUserStatistics(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	//
	//	// Get user from context
	//	user := ctx.GetUser()
	//	if user == nil {
	//		ctx.Unauthorized("missing user, please login first")
	//		return
	//	}
	//
	//	// Get token from URL query parameter
	//	var token *common.Token
	//	tokenStr := req.URL.Query().Get("token")
	//
	//	if tokenStr != "" {
	//		for _, t := range user.Tokens {
	//			if t.Token == tokenStr {
	//				token = t
	//				break
	//			}
	//		}
	//		if token == nil {
	//			ctx.InvalidParameter("token")
	//			return
	//		}
	//	}
	//
	//	// Get server statistics
	//	stats, err := ctx.GetMetadataBackend().GetUserStatistics(user, token)
	//	if err != nil {
	//		ctx.InternalServerError("unable to get user statistics", err)
	//		return
	//	}
	//
	//	// Print stats in the json response.
	//	var json []byte
	//	if json, err = utils.ToJson(stats); err != nil {
	//		panic(fmt.Errorf("unable to serialize json response : %s", err))
	//	}
	//
	//	_, _ = resp.Write(json)
}

func getUserAndToken(ctx *context.Context, req *http.Request) (user *common.User, token *common.Token, err error) {
	// Get user from context
	user = ctx.GetUser()
	if user == nil {
		return nil, nil, common.NewHTTPError("missing user, please login first", nil, http.StatusUnauthorized)
	}

	// Get token from URL query parameter
	tokenStr := req.URL.Query().Get("token")
	if tokenStr != "" {
		token, err = ctx.GetMetadataBackend().GetToken(tokenStr)
		if err != nil {
			ctx.InternalServerError("unable to get token", err)
			return nil, nil, common.NewHTTPError("unable to get token", err, http.StatusInternalServerError)
		}
		if token == nil {
			return nil, nil, common.NewHTTPError("token not found", nil, http.StatusNotFound)
		}
		if token.UserID != user.ID {
			return nil, nil, common.NewHTTPError("token not found", nil, http.StatusNotFound)
		}
	}

	return user, token, nil
}
