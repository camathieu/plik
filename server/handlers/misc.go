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
	"image/png"
	"net/http"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// GetVersion return the build information.
func GetVersion(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Print version and build information in the json response.
	json, err := utils.ToJson(common.GetBuildInfo())
	if err != nil {
		log.Warningf("Unable to serialize json response : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
		return
	}

	resp.Write(json)
}

// GetConfiguration return the server configuration
func GetConfiguration(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	// Print configuration in the json response.
	json, err := utils.ToJson(config)
	if err != nil {
		log.Warningf("Unable to serialize response body : %s", err)
		context.Fail(ctx, req, resp, "Unable to serialize response body", http.StatusInternalServerError)
		return
	}
	resp.Write(json)
}

// Logout return the server configuration
func Logout(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	common.Logout(resp)
}

// GetQrCode return a QRCode for the requested URL
func GetQrCode(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)

	// Check params
	urlParam := req.FormValue("url")
	sizeParam := req.FormValue("size")

	// Parse int on size
	sizeInt, err := strconv.Atoi(sizeParam)
	if err != nil {
		sizeInt = 250
	}
	if sizeInt > 1000 {
		log.Warning("QRCode size must be lower than 1000")
		context.Fail(ctx, req, resp, "QRCode size must be lower than 1000", http.StatusBadRequest)
		return
	}
	if sizeInt <= 0 {
		log.Warning("QRCode size must be positive")
		context.Fail(ctx, req, resp, "QRCode size must be positive", http.StatusBadRequest)
		return
	}

	// Generate QRCode png from url
	qrcode, err := qr.Encode(urlParam, qr.H, qr.Auto)
	if err != nil {
		log.Warningf("Unable to generate QRCode : %s", err)
		context.Fail(ctx, req, resp, "Unable to generate QRCode", http.StatusInternalServerError)
		return
	}

	// Scale QRCode png size
	qrcode, err = barcode.Scale(qrcode, sizeInt, sizeInt)
	if err != nil {
		log.Warningf("Unable to scale QRCode : %s", err)
		context.Fail(ctx, req, resp, "Unable to generate QRCode", http.StatusInternalServerError)
		return
	}

	resp.Header().Add("Content-Type", "image/png")
	err = png.Encode(resp, qrcode)
	if err != nil {
		log.Warningf("Unable to encode png : %s", err)
	}
}

// File status MUST be removed before to call this
func DeleteRemovedFile(ctx *juliet.Context, upload *common.Upload, file *common.File) (err error) {
	backend := context.GetDataBackend(ctx)
	err = backend.RemoveFile(upload, file.ID)
	if err != nil {
		return fmt.Errorf("error while deleting file %s (%s) from upload %s : %s", file.Name, file.ID, upload.ID, err)
	}

	tx := func(u *common.Upload) error {
		if u == nil {
			return fmt.Errorf("missing upload from transaction")
		}

		f, ok := u.Files[file.ID]
		if !ok {
			return fmt.Errorf("unable to find file %s (%s)", file.Name, file.ID)
		}
		if f.Status != common.FILE_REMOVED {
			return fmt.Errorf("file %s (%s) status is not %s", file.Name, file.ID, common.FILE_REMOVED)
		}
		f.Status = common.FILE_DELETED

		return nil
	}

	upload, err = context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
	if err != nil {
		return fmt.Errorf("Unable to update upload metadata : %s", err)
	}

	// Remove upload if no files anymore
	RemoveEmptyUpload(ctx, upload)

	return nil
}

// RemoveUploadIfNoFileAvailable iterates on upload files and remove upload files
// and metadata if all the files have been downloaded (useful for OneShot uploads)
func RemoveEmptyUpload(ctx *juliet.Context, upload *common.Upload) {
	log := context.GetLogger(ctx)

	// Test if there are remaining files
	filesInUpload := len(upload.Files)
	for _, f := range upload.Files {
		if f.Status == common.FILE_DELETED {
			filesInUpload--
		}
	}

	if filesInUpload == 0 {
		err := context.GetMetadataBackend(ctx).RemoveUpload(upload)
		if err != nil {
			log.Warningf("Unable to remove upload : %s", err)
			return
		}
	}

	return
}
