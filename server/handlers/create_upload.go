package handlers

import (
	"encoding/base64"
	"fmt"
	"github.com/root-gg/plik/server/common"
	"io/ioutil"
	"net/http"

	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// CreateUpload create a new upload
func CreateUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()
	config := ctx.GetConfig()

	if !ctx.IsWhitelisted() {
		ctx.Forbidden("untrusted source IP address")
		return
	}

	// TODO NOT A GOOD IDEA HERE !!!

	// Read request body
	defer func() { _ = req.Body.Close() }()
	req.Body = http.MaxBytesReader(resp, req.Body, 1048576)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ctx.BadRequest("unable to read request body", err)
		return
	}

	// Create upload
	upload := &common.Upload{}

	// Deserialize json body
	version := 0
	if len(body) > 0 {
		version, err = common.UnmarshalUpload(body, upload)
		if err != nil {
			ctx.BadRequest("unable to unserialize upload : %s", err.Error())
		}
	}

	// Assign context parameters ( ip / user / token )
	ctx.SetUploadContext(upload)

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, upload.ID)
	log.SetPrefix(prefix)
	ctx.SetUpload(upload)

	// Protect upload with HTTP basic auth
	// Add Authorization header to the response for convenience
	// So clients can just copy this header into the next request
	if upload.Password != "" {
		if upload.Login == "" {
			upload.Login = "plik"
		}

		// The Authorization header will contain the base64 version of "login:password"
		// Save only the md5sum of this string to authenticate further requests
		b64str := base64.StdEncoding.EncodeToString([]byte(upload.Login + ":" + upload.Password))
		upload.Password, err = utils.Md5sum(b64str)
		if err != nil {
			ctx.BadRequest("unable to generate password hash : %s", err)
			return
		}
		resp.Header().Add("Authorization", "Basic "+b64str)
	}

	// Set and validate upload parameters
	err = upload.PrepareInsert(config)
	if err != nil {
		ctx.BadRequest(err.Error())
		return
	}

	// Save the metadata
	err = ctx.GetMetadataBackend().CreateUpload(upload)
	if err != nil {
		ctx.InternalServerError("create upload error", err)
		return
	}

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	uploadToken := upload.UploadToken
	upload.Sanitize()
	upload.DownloadDomain = config.DownloadDomain

	// Show upload token since its an upload creation
	upload.UploadToken = uploadToken
	upload.IsAdmin = true

	// Print upload metadata in the json response.
	bytes, err := common.MarshalUpload(upload, version)
	if err != nil {
		ctx.InternalServerError("unable to serialize upload", err)
	}

	_, _ = resp.Write(bytes)
}
