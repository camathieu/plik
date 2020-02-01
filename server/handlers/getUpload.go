package handlers

import (
	"fmt"
	"net/http"

	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// GetUpload return upload metadata
func GetUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	// Get upload from context
	upload := ctx.GetUpload()

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	upload.Sanitize()
	upload.DownloadDomain = config.DownloadDomain

	if ctx.IsUploadAdmin() {
		upload.Admin = true
	}

	// Print upload metadata in the json response.
	json, err := utils.ToJson(upload)
	if err != nil {
		ctx.InternalServerError(fmt.Errorf("unable to serialize json response : %s", err))
		return
	}

	_, _ = resp.Write(json)
}
