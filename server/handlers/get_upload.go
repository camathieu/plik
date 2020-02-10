package handlers

import (
	"github.com/root-gg/utils"
	"net/http"

	"github.com/root-gg/plik/server/context"
)

// GetUpload return upload metadata
func GetUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	// Get upload from context
	upload := ctx.GetUpload()
	if upload == nil {
		panic("missing upload from context")
	}

	files, err := ctx.GetMetadataBackend().GetFiles(upload)
	if err != nil {
		ctx.InternalServerError("unable to get upload files", err)
		return
	}

	upload.Files = files

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	upload.Sanitize()
	upload.DownloadDomain = config.DownloadDomain

	if ctx.IsUploadAdmin() {
		upload.IsAdmin = true
	}

	// Print upload metadata in the json response.
	bytes, err := utils.ToJson(upload)
	if err != nil {
		ctx.InternalServerError("unable serialize json response", err)
		return
	}

	_, _ = resp.Write(bytes)
}
