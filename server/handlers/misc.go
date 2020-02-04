package handlers

import (
	"fmt"
	"image/png"
	"net/http"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// GetVersion return the build information.
func GetVersion(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	// Print version and build information in the json response.
	json, err := utils.ToJson(common.GetBuildInfo())
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// GetConfiguration return the server configuration
func GetConfiguration(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	config := ctx.GetConfig()

	// Print configuration in the json response.
	json, err := utils.ToJson(config)
	if err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(json)
}

// Logout return the server configuration
func Logout(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	common.Logout(resp)
}

// GetQrCode return a QRCode for the requested URL
func GetQrCode(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	// Check params
	urlParam := req.FormValue("url")
	sizeParam := req.FormValue("size")

	// Parse int on size
	sizeInt, err := strconv.Atoi(sizeParam)
	if err != nil {
		sizeInt = 250
	}
	if sizeInt <= 0 {
		ctx.BadRequest("QRCode size must be positive")
		return
	}
	if sizeInt > 1000 {
		ctx.BadRequest("QRCode size must be lower than 1000")
		return
	}

	// Generate QRCode png from url
	qrcode, err := qr.Encode(urlParam, qr.H, qr.Auto)
	if err != nil {
		ctx.InternalServerError("unable to generate QRCode", err)
		return
	}

	// Scale QRCode png size
	qrcode, err = barcode.Scale(qrcode, sizeInt, sizeInt)
	if err != nil {
		ctx.InternalServerError("unable to scale QRCode : %s", err)
		return
	}

	resp.Header().Add("Content-Type", "image/png")
	err = png.Encode(resp, qrcode)
	if err != nil {
		ctx.InternalServerError("unable to encore png : %s", err)
		return
	}
}

// DeleteRemovedFile deletes a removed file
func DeleteRemovedFile(ctx *context.Context, upload *common.Upload, file *common.File) (err error) {

	if upload == nil {
		return fmt.Errorf("upload parameter is nil")
	}

	if file == nil {
		return fmt.Errorf("file parameter is nil")
	}

	// /!\ File status MUST be removed before to call this /!\

	backend := ctx.GetDataBackend()
	err = backend.RemoveFile(upload, file.ID)
	if err != nil {
		return fmt.Errorf("error while deleting file %s (%s) from upload %s : %s", file.Name, file.ID, upload.ID, err)
	}

	tx := func(u *common.Upload) error {
		if u == nil {
			return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
		}

		f, ok := u.Files[file.ID]
		if !ok {
			return fmt.Errorf("unable to find file %s (%s)", file.Name, file.ID)
		}
		if f.Status != common.FileRemoved {
			return fmt.Errorf("invalid file %s (%s) status %s, expected %s", file.Name, file.ID, f.Status, common.FileRemoved)
		}
		f.Status = common.FileDeleted

		return nil
	}

	upload, err = ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		return fmt.Errorf("unable to update upload metadata : %s", err)
	}

	// Remove upload if no files anymore
	//RemoveEmptyUpload(ctx, upload)

	return nil
}

//// RemoveEmptyUpload iterates on upload files and remove upload files
//// and metadata if all the files have been downloaded (useful for OneShot uploads)
//func RemoveEmptyUpload(ctx *context.Context, upload *common.Upload) {
//	log := ctx.GetLogger()
//
//	// Test if there are remaining files
//	filesInUpload := len(upload.Files)
//	for _, f := range upload.Files {
//		if f.Status == common.FileDeleted {
//			filesInUpload--
//		}
//	}
//
//	if filesInUpload == 0 {
//		err := ctx.GetMetadataBackend().RemoveUpload(upload)
//		if err != nil {
//			log.Warningf("unable to remove upload : %s", err)
//			return
//		}
//	}
//
//	return
//}

// If a download domain is specified verify that the request comes from this specific domain
func checkDownloadDomain(ctx *context.Context) bool {
	log := ctx.GetLogger()
	config := ctx.GetConfig()
	req := ctx.GetReq()
	resp := ctx.GetResp()

	if config.GetDownloadDomain() != nil {
		if req.Host != config.GetDownloadDomain().Host {
			downloadURL := fmt.Sprintf("%s://%s%s",
				config.GetDownloadDomain().Scheme,
				config.GetDownloadDomain().Host,
				req.RequestURI)
			log.Warningf("invalid download domain %s, expected %s", req.Host, config.GetDownloadDomain().Host)
			http.Redirect(resp, req, downloadURL, http.StatusMovedPermanently)
			return false
		}
	}

	return true
}

func handleTxError(ctx *context.Context, message string, err error) {
	if txError, ok := err.(common.HTTPError); ok {
		ctx.Fail(txError.Error(), nil, txError.GetStatusCode())
	} else {
		ctx.InternalServerError(message, err)
	}
}
