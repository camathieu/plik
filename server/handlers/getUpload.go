package handlers

import (
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// GetUpload return upload metadata
func GetUpload(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	// Get upload from context
	upload := context.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in getUploadHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	upload.Sanitize()
	upload.DownloadDomain = config.DownloadDomain

	if context.IsUploadAdmin(ctx) {
		upload.Admin = true
	}

	// Print upload metadata in the json response.
	json, err := utils.ToJson(upload)
	if err != nil {
		log.Warningf("Unable to serialize json response : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
		return
	}

	resp.Write(json)
}
