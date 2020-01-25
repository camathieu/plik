/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

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
		context.Fail(ctx, req, resp, "Internal error", 500)
		return
	}

	// Check authorization
	if !upload.Removable && !context.IsUploadAdmin(ctx) {
		log.Warningf("Unable to remove file : unauthorized")
		context.Fail(ctx, req, resp, "You are not allowed to remove file from this upload", 403)
		return
	}

	// Get file from context
	file := context.GetFile(ctx)
	if file == nil {
		// This should never append
		log.Critical("Missing file in removeFileHandler")
		context.Fail(ctx, req, resp, "Internal error", 500)
		return
	}

	// Check if file is not already removed
	if file.Status == "removed" {
		log.Warning("Can't remove an already removed file")
		context.Fail(ctx, req, resp, fmt.Sprintf("File %s has already been removed", file.Name), 404)
		return
	}

	remove := true
	tx := func(u *common.Upload) error {
		f, ok := u.Files[file.ID]
		if !ok {
			return fmt.Errorf("Unable to find file %s", file.ID)
		}
		if f.Status != common.FILE_UPLOADED {
			// Nothing to do
			remove = false
			return nil
		}
		f.Status = common.FILE_REMOVED
		return nil
	}

	err := context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
	if err != nil {
		log.Warningf("Unable to update upload %s", upload.ID)
		context.Fail(ctx, req, resp, "Unable to add file", http.StatusInternalServerError)
		return
	}

	if remove {
		err := DeleteFile(ctx, upload, file)
		if err != nil {
			log.Warningf("Unable to delete file : %s")
			// Do not block here
		}

		// Remove upload if no files anymore
		RemoveEmptyUpload(ctx, upload)
	}

	resp.Write([]byte("ok"))
}
