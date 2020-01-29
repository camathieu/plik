package handlers

import (
	"fmt"
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// RemoveFile remove a file from an existing upload
func RemoveFile(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get upload from context
	upload := context.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in removeFileHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check authorization
	if !upload.Removable && !context.IsUploadAdmin(ctx) {
		log.Warningf("Unable to remove file : unauthorized")
		context.Fail(ctx, req, resp, "You are not allowed to remove file from this upload", http.StatusForbidden)
		return
	}

	// Get file from context
	file := context.GetFile(ctx)
	if file == nil {
		// This should never append
		log.Critical("Missing file in removeFileHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check if file is not already removed
	if file.Status == "removed" {
		log.Warning("Can't remove an already removed file")
		context.Fail(ctx, req, resp, fmt.Sprintf("File %s has already been removed", file.Name), http.StatusNotFound)
		return
	}

	remove := true
	tx := func(u *common.Upload) error {
		if u == nil {
			return fmt.Errorf("missing upload from transaction")
		}

		f, ok := u.Files[file.ID]
		if !ok {
			return common.NewTxError(fmt.Sprintf("File %s (%s) not found", file.Name, file.ID), http.StatusNotFound)
		}

		if f.Status == common.FileRemoved || f.Status == common.FileDeleted {
			// Nothing to do
			remove = false
			return nil
		}

		f.Status = common.FileRemoved
		return nil
	}

	upload, err := context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
	if err != nil {
		if txError, ok := err.(common.TxError); ok {
			context.Fail(ctx, req, resp, txError.Error(), txError.GetStatusCode())
		} else {
			log.Warningf("Unable to update upload : %s", err)
			context.Fail(ctx, req, resp, "Unable to remove file", http.StatusInternalServerError)
		}
		return
	}

	if remove {
		err := DeleteRemovedFile(ctx, upload, file)
		if err != nil {
			log.Warningf("Unable to delete file %s (%s) : %s", file.Name, file.ID, err)
			// Do not block here
		}
	}

	_, _ = resp.Write([]byte("ok"))
}
