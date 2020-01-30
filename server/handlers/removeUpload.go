package handlers

import (
	"github.com/root-gg/plik/server/common"
	"net/http"


	"github.com/root-gg/plik/server/context"
)

// RemoveUpload remove an upload and all associated files
func RemoveUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()

	// Get upload from context
	upload := ctx.GetUpload()
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in removeUploadHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check authorization
	if !upload.Removable && !ctx.IsUploadAdmin() {
		log.Warningf("Unable to remove upload : unauthorized")
		context.Fail(ctx, req, resp, "You are not allowed to remove this upload", http.StatusForbidden)
		return
	}

	var files []*common.File
	tx := func(u *common.Upload) error {
		files = []*common.File{}
		for _, f := range u.Files {
			if f.Status == common.FileUploaded {
				files = append(files, f)
			}

			if f.Status != common.FileDeleted {
				f.Status = common.FileRemoved
			}
		}
		return nil
	}

	upload, err := ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		log.Warningf("Unable to update upload metadata : %s", err)
		context.Fail(ctx, req, resp, "Unable to update upload metadata", http.StatusInternalServerError)
		return
	}

	// Remove files
	for _, file := range upload.Files {
		err = DeleteRemovedFile(ctx, upload, file)
		if err != nil {
			// Don't block here
			log.Warningf("Unable to delete file %s (%s) : %s", file.Name, file.ID, err)
		}
	}

	if err != nil {
		context.Fail(ctx, req, resp, "Unable to remove all files", http.StatusInternalServerError)
		return
	}
}
