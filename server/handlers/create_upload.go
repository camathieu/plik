package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// CreateUpload create a new upload
func CreateUpload(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()
	config := ctx.GetConfig()

	user := ctx.GetUser()
	if user == nil {
		if config.NoAnonymousUploads {
			ctx.BadRequest("anonymous uploads are disabled, please authenticate first")
			return
		} else if !ctx.IsWhitelisted() {
			ctx.Forbidden("untrusted source IP address")
			return
		}
	}

	upload := common.NewUpload()

	// Read request body
	defer func() { _ = req.Body.Close() }()
	req.Body = http.MaxBytesReader(resp, req.Body, 1048576)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ctx.BadRequest("unable to read request body", err)
		return
	}

	// Deserialize json body
	if len(body) > 0 {
		err = json.Unmarshal(body, upload)
		if err != nil {
			ctx.BadRequest("unable to deserialize request body : %s", err)
			return
		}
	}

	// Limit number of files per upload
	if len(upload.Files) > config.MaxFilePerUpload {
		ctx.BadRequest("too many files. maximum is %d", config.MaxFilePerUpload)
		return
	}

	// Set upload id, creation date, upload token, ...
	upload.Create()

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, upload.ID)
	log.SetPrefix(prefix)
	ctx.SetUpload(upload)

	// Set upload remote IP
	upload.RemoteIP = ctx.GetSourceIP().String()

	// Set upload user and token
	if user != nil {
		upload.User = user.ID
		token := ctx.GetToken()
		if token != nil {
			token := ctx.GetToken()
			if token != nil {
				upload.Token = token.Token
			}
		}
	}

	if upload.OneShot && !config.OneShot {
		ctx.BadRequest("one shot downloads are not enabled")
		return
	}

	if upload.Removable && !config.Removable {
		ctx.BadRequest("removable uploads are not enabled")
		return
	}

	if upload.Stream && !config.StreamMode {
		ctx.BadRequest("stream mode is not enabled")
		return
	}

	// TTL = Time in second before the upload expiration
	// 0 	-> No ttl specified : default value from configuration
	// -1	-> No expiration : checking with configuration if that's ok
	switch upload.TTL {
	case 0:
		upload.TTL = config.DefaultTTL
	case -1:
		if config.MaxTTL != -1 {
			ctx.BadRequest("cannot set infinite ttl (maximum allowed is : %d)", config.MaxTTL)
			return
		}
	default:
		if upload.TTL <= 0 {
			ctx.InvalidParameter("ttl")
			return
		}
		if config.MaxTTL > 0 && upload.TTL > config.MaxTTL {
			ctx.InvalidParameter("ttl. (maximum allowed is : %d)", config.MaxTTL)
			return
		}
	}

	// Protect upload with HTTP basic auth
	// Add Authorization header to the response for convenience
	// So clients can just copy this header into the next request
	if upload.Password != "" {
		if !config.ProtectedByPassword {
			ctx.BadRequest("password protection is not enabled")
			return
		}

		upload.ProtectedByPassword = true
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

	// Check the token validity with api.yubico.com
	// Only the Yubikey id part of the token is stored
	// The yubikey id is the 12 first characters of the token
	// The 32 lasts characters are the actual OTP
	if upload.Yubikey != "" {
		upload.ProtectedByYubikey = true

		if !config.YubikeyEnabled {
			ctx.BadRequest("yubikey are disabled on this server")
			return
		}

		_, ok, err := config.GetYubiAuth().Verify(upload.Yubikey)
		if err != nil {
			ctx.InternalServerError("unable to validate yubikey token : %s", err)
			return
		}

		if !ok {
			ctx.InvalidParameter("yubikey token")
			return
		}

		upload.Yubikey = upload.Yubikey[:12]
	}

	// Create files
	for i, file := range upload.Files {

		// Check file name length
		if len(file.Name) > 1024 {
			ctx.InvalidParameter("file name, at least one file name is too long, maximum length is 1024 characters")
			return
		}

		file.GenerateID()
		file.Status = "missing"
		delete(upload.Files, i)
		upload.Files[file.ID] = file
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
	upload.Admin = true

	// Print upload metadata in the json response.
	var bytes []byte
	if bytes, err = utils.ToJson(upload); err != nil {
		panic(fmt.Errorf("unable to serialize json response : %s", err))
	}

	_, _ = resp.Write(bytes)
}
