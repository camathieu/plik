package handlers

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/utils"
	"net/http"

	"github.com/root-gg/plik/server/context"
)

// GetUsers return users
func GetUsers(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Check authorization
	if !ctx.IsAdmin() {
		ctx.Forbidden("you need administrator privileges")
		return
	}

	pagingQuery := ctx.GetPagingQuery()
	if pagingQuery == nil {
		ctx.InternalServerError("missing paging query", nil)
		return
	}

	// Get uploads
	users, cursor, err := ctx.GetMetadataBackend().GetUsers("", pagingQuery)
	if err != nil {
		ctx.InternalServerError("unable to get users : %s", err)
		return
	}

	pr := common.NewPagingResponse(users, cursor)

	json, err := utils.ToJson(pr)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// GetServerStatistics return the server statistics
func GetServerStatistics(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {

	// Check authorization
	if !ctx.IsAdmin() {
		ctx.Forbidden("you need administrator privileges")
		return
	}

	// Get server statistics
	stats, err := ctx.GetMetadataBackend().GetServerStatistics()
	if err != nil {
		ctx.InternalServerError("unable to get server statistics : %s", err)
		return
	}

	// Print stats in the json response.
	json, err := utils.ToJson(stats)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}
