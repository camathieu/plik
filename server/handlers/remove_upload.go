package handlers

import (
	"net/http"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// RemoveUpload remove an upload and all associated files
func RemoveUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
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
	for _, file := range upload.Files {
		if file.Status == common.FileUploaded {
			files = append(files, file)
		}

		if file.Status == common.FileRemoved || file.Status == common.FileDeleted {
			continue
		}

		currentStatus := file.Status
		file.Status = common.FileRemoved
		err := ctx.GetMetadataBackend().AddOrUpdateFile(upload, file, currentStatus)
		if err != nil {
			ctx.InternalServerError("unable to update file metadata", err)
			return
		}

		if currentStatus == common.FileUploaded {
			err = DeleteRemovedFile(ctx, upload, file)
			if err != nil {
				ctx.InternalServerError("unable to remove file", err)
				return
			}
		}
	}

	err := ctx.GetMetadataBackend().RemoveUpload(upload)
	if err != nil {
		ctx.InternalServerError("unable to remove upload metadata", err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}
