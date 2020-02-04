package handlers

import (
	"fmt"
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
		ctx.InternalServerError("missing upload from context", nil)
		return
	}

	// Check authorization
	if !upload.Removable && !ctx.IsUploadAdmin() {
		ctx.Forbidden("you are not allowed to remove this upload")
		return
	}

	var files []*common.File
	tx := func(u *common.Upload) error {
		if u == nil {
			return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
		}

		files = []*common.File{}
		for _, f := range u.Files {
			if f.Status == common.FileUploading {
				return common.NewHTTPError(fmt.Sprintf("file %s (%s) is still uploading", f.Name, f.ID), http.StatusBadRequest)
			}

			if f.Status == common.FileUploaded || f.Status == common.FileRemoved {
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
		handleTxError(ctx, "unable to update upload metadata", err)
		return
	}

	// Remove files
	for _, file := range upload.Files {
		err = DeleteRemovedFile(ctx, upload, file)
		if err != nil {
			// Don't block here
			log.Warningf("unable to delete file %s (%s) : %s", file.Name, file.ID, err)
		}
	}
	if err != nil {
		ctx.InternalServerError("unable to remove all files", err)
		return
	}

	// Reload upload metadata
	upload, err = ctx.GetMetadataBackend().GetUpload(upload.ID)
	if err != nil {
		ctx.InternalServerError("unable to reload upload metadata", err)
	} else if upload == nil {
		ctx.NotFound("upload does not exist anymore")
	}

	// Test if there are remaining files
	filesInUpload := len(upload.Files)
	for _, f := range upload.Files {
		if f.Status == common.FileDeleted {
			filesInUpload--
		}
	}

	if filesInUpload > 0 {
		ctx.BadRequest("there are still %d files in this upload", filesInUpload)
	}

	err = ctx.GetMetadataBackend().RemoveUpload(upload)
	if err != nil {
		ctx.InternalServerError("unable to reload upload metadata", err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}
