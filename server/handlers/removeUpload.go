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
	"github.com/root-gg/plik/server/common"
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
)

// RemoveUpload create a new upload
func RemoveUpload(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Get upload from context
	upload := context.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in removeUploadHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check authorization
	if !upload.Removable && !context.IsUploadAdmin(ctx) {
		log.Warningf("Unable to remove upload : unauthorized")
		context.Fail(ctx, req, resp, "You are not allowed to remove this upload", http.StatusForbidden)
		return
	}

	var files []*common.File
	tx := func(u *common.Upload) error {
		files = []*common.File{}
		for _, f := range u.Files {
			if f.Status == common.FILE_UPLOADED {
				files = append(files, f)
			}

			if f.Status != common.FILE_DELETED {
				f.Status = common.FILE_REMOVED
			}
		}
		return nil
	}

	upload, err := context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
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
			log.Warningf("Unable to delete file %s (%s)", file.Name, file.ID)
		}
	}

	if err != nil {
		context.Fail(ctx, req, resp, "Unable to remove all files", http.StatusInternalServerError)
		return
	}
}
