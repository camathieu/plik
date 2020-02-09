package handlers

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
	"net/http"
)

// UserInfo return user information ( name / email / tokens / ... )
func UserInfo(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	user.IsAdmin = ctx.IsAdmin()

	// Serialize user to JSON
	// Print token in the json response.
	json, err := utils.ToJson(user)
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

	err := ctx.GetMetadataBackend().DeleteUser(user)
	if err != nil {
		ctx.InternalServerError("unable to remove user", err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}

// GetUserUploads get user uploads
func GetUserUploads(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	//log := ctx.GetLogger()
	//
	//// Get user from context
	//user := ctx.GetUser()
	//if user == nil {
	//	ctx.Unauthorized("missing user, please login first")
	//	return
	//}
	//
	//// Get token from URL query parameter
	//var token *common.Token
	//tokenStr := req.URL.Query().Get("token")
	//
	//if tokenStr != "" {
	//	for _, t := range user.Tokens {
	//		if t.Token == tokenStr {
	//			token = t
	//			break
	//		}
	//	}
	//	if token == nil {
	//		ctx.InvalidParameter("token")
	//		return
	//	}
	//}
	//
	//// Get uploads
	//ids, err := ctx.GetMetadataBackend().GetUserUploads(user, token)
	//if err != nil {
	//	ctx.InternalServerError("unable to get user uploads : %s", err)
	//	return
	//}
	//
	//// Get size from URL query parameter
	//size := 100
	//sizeStr := req.URL.Query().Get("size")
	//if sizeStr != "" {
	//	size, err = strconv.Atoi(sizeStr)
	//	if err != nil || size <= 0 || size > 100 {
	//		ctx.InvalidParameter("size, must be positive integer less than 100")
	//		return
	//	}
	//}
	//
	//// Get offset from URL query parameter
	//offset := 0
	//offsetStr := req.URL.Query().Get("offset")
	//if offsetStr != "" {
	//	offset, err = strconv.Atoi(offsetStr)
	//	if err != nil || offset < 0 {
	//		ctx.InvalidParameter("offset, must be positive integer")
	//		return
	//	}
	//}
	//
	//// Adjust offset
	//if offset > len(ids) {
	//	offset = len(ids)
	//}
	//
	//// Adjust size
	//if offset+size > len(ids) {
	//	size = len(ids) - offset
	//}
	//
	//var uploads []*common.Upload
	//for _, id := range ids[offset : offset+size] {
	//	upload, err := ctx.GetMetadataBackend().GetUpload(id)
	//	if err != nil {
	//		log.Warningf("Unable to get upload %s : %s", id, err)
	//		continue
	//	}
	//	if upload == nil {
	//		log.Warningf("Upload %s not found", id)
	//		continue
	//	}
	//
	//	if !upload.IsExpired() {
	//		token := upload.Token
	//		upload.Sanitize()
	//		upload.Token = token
	//		upload.Admin = true
	//		uploads = append(uploads, upload)
	//	}
	//}
	//
	//// Print uploads in the json response.
	//var json []byte
	//if json, err = utils.ToJson(uploads); err != nil {
	//	panic(fmt.Errorf("unable to serialize json response : %s", err))
	//}
	//
	//_, _ = resp.Write(json)
}

// RemoveUserUploads delete all user uploads
func RemoveUserUploads(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	// Get user from context
	user := ctx.GetUser()
	if user == nil {
		ctx.Unauthorized("missing user, please login first")
		return
	}

	// Get token from URL query parameter
	var token *common.Token
	tokenStr := req.URL.Query().Get("token")
	if tokenStr != "" {
		for _, t := range user.Tokens {
			if t.Token == tokenStr {
				token = t
			}
		}
		if token == nil {
			ctx.InvalidParameter("token")
			return
		}
	}

	// Delete user uploads
	deleted := 0
	f := func(upload *common.Upload) (err error) {
		deleted++
		return ctx.GetMetadataBackend().DeleteUpload(upload)
	}

	err := ctx.GetMetadataBackend().ForEachUserUpload(user, token, f)
	if err != nil {
		ctx.InternalServerError("unable to delete user uploads : %s", err)
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
