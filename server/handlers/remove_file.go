package handlers

import (
	"fmt"
	"net/http"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// RemoveFile remove a file from an existing upload
func RemoveFile(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()

	// Get upload from context
	upload := ctx.GetUpload()
	if upload == nil {
		ctx.InternalServerError("missing upload from context", nil)
		return
	}

	// Check authorization
	if !upload.Removable && !ctx.IsUploadAdmin() {
		ctx.Forbidden("you are not allowed to remove files from this upload")
		return
	}

	// Get file from context
	file := ctx.GetFile()
	if file == nil {
		ctx.InternalServerError("missing file from context", nil)
		return
	}

	// Check if file is not already removed
	if file.Status != common.FileUploaded {
		ctx.NotFound("file %s (%s) is not removable : %s", file.Name, file.ID, file.Status)
		return
	}

	remove := true
	tx := func(u *common.Upload) error {
		if u == nil {
			return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
		}

		f, ok := u.Files[file.ID]
		if !ok {
			return common.NewHTTPError(fmt.Sprintf("file %s (%s) is not available anymore", file.Name, file.ID), http.StatusNotFound)
		}

		if f.Status == common.FileRemoved || f.Status == common.FileDeleted {
			// Nothing to do
			remove = false
			return nil
		}

		f.Status = common.FileRemoved
		return nil
	}

	upload, err := ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		handleTxError(ctx, "unable to update upload metadata", err)
		return
	}

	// Get updated file
	var ok bool
	file, ok = upload.Files[file.ID]
	if !ok {
		ctx.InternalServerError("missing file from upload after metadata update", nil)
		return
	}

	if remove {
		err := DeleteRemovedFile(ctx, upload, file)
		if err != nil {
			log.Warningf("unable to delete file %s (%s) : %s", file.Name, file.ID, err)
		}
	}

	_, _ = resp.Write([]byte("ok"))
}
