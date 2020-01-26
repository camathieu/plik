/*
 * Charles-Antoine Mathieu <charles-antoine.mathieu@ovh.net>
 */

package handlers

import (
	"net/http"
	"strconv"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// GetUsers return users information ( name / email / tokens / ... )
func GetUsers(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get user from context
	user := context.GetUser(ctx)
	if user == nil {
		context.Fail(ctx, req, resp, "Missing user, Please login first", http.StatusUnauthorized)
		return
	}

	if !context.IsAdmin(ctx) {
		context.Fail(ctx, req, resp, "You need administrator privileges", http.StatusForbidden)
		return
	}

	ids, err := context.GetMetadataBackend(ctx).GetUsers()
	if err != nil {
		log.Warningf("Unable to get users : %s", err)
		context.Fail(ctx, req, resp, "Unable to get users", http.StatusInternalServerError)
		return
	}

	// Get size from URL query parameter
	size := 100
	sizeStr := req.URL.Query().Get("size")
	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size <= 0 || size > 1000 {
			log.Warningf("Invalid size parameter : %s", sizeStr)
			context.Fail(ctx, req, resp, "Invalid size parameter", http.StatusBadRequest)
			return
		}
	}

	// Get offset from URL query parameter
	offset := 0
	offsetStr := req.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			log.Warningf("Invalid offset parameter : %s", offsetStr)
			context.Fail(ctx, req, resp, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
	}

	// Adjust offset
	if offset > len(ids) {
		offset = len(ids)
	}

	// Adjust size
	if offset+size > len(ids) {
		size = len(ids) - offset
	}

	var users []*common.User
	for _, id := range ids[offset : offset+size] {
		user, err := context.GetMetadataBackend(ctx).GetUser(id)
		if err != nil {
			log.Warningf("Unable to get user %s : %s", id, err)
			continue
		}

		// Remove tokens
		user.Tokens = nil

		users = append(users, user)
	}

	// Print users in the json response.
	var json []byte
	if json, err = utils.ToJson(users); err != nil {
		log.Warningf("Unable to serialize json response : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
		return
	}

	resp.Write(json)
}

// GetServerStatistics return the server statistics
func GetServerStatistics(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get user from context
	user := context.GetUser(ctx)
	if user == nil {
		context.Fail(ctx, req, resp, "Missing user, Please login first", http.StatusUnauthorized)
		return
	}

	if !context.IsAdmin(ctx) {
		context.Fail(ctx, req, resp, "You need administrator privileges", http.StatusForbidden)
		return
	}

	// Get server statistics
	stats, err := context.GetMetadataBackend(ctx).GetServerStatistics()
	if err != nil {
		log.Warningf("Unable to get server statistics : %s", err)
		context.Fail(ctx, req, resp, "Unable to get server statistics", http.StatusInternalServerError)
		return
	}

	// Print stats in the json response.
	var json []byte
	if json, err = utils.ToJson(stats); err != nil {
		log.Warningf("Unable to serialize json response : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
		return
	}

	resp.Write(json)
}
