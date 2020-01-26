package context

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/root-gg/juliet"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/metadata"
	"github.com/stretchr/testify/require"
)

// declare context keys
type key string

var configKey = "config"
var loggerKey = "logger"
var metadataBackendKey = "metadata_backend"
var dataBackendKey = "data_backend"
var streamBackendKey = "stream_backend"
var sourceIPKey = "source_ip"
var whitelistedKey = "is_whitelisted"
var userKey = "user"
var tokenKey = "token"
var uploadKey = "upload"
var fileKey = "file"
var adminKey = "is_amdin"
var uploadAdminKey = "is_upload_admin"
var redirectOnFailureKey = "is_redirect_on_failure"
var quickKey = "is_quick"

// SetConfig set config for the context
func SetConfig(ctx *juliet.Context, config *common.Configuration) {
	ctx.Set(configKey, config)

}

// GetConfig from the request context.
func GetConfig(ctx *juliet.Context) (config *common.Configuration) {
	if config, ok := ctx.Get(configKey); ok {
		return config.(*common.Configuration)
	}
	return nil
}

func SetLogger(ctx *juliet.Context, log *logger.Logger) {
	ctx.Set(loggerKey, log)
}

// GetLogger from the request context.
func GetLogger(ctx *juliet.Context) *logger.Logger {
	if log, ok := ctx.Get(loggerKey); ok {
		return log.(*logger.Logger)
	}
	return nil
}

func SetMetadataBackend(ctx *juliet.Context, backend metadata.Backend) {
	ctx.Set(metadataBackendKey, backend)
}

// GetMetadataBackend from the request context.
func GetMetadataBackend(ctx *juliet.Context) metadata.Backend {
	if backend, ok := ctx.Get(metadataBackendKey); ok {
		return backend.(metadata.Backend)
	}
	return nil
}

func SetDataBackend(ctx *juliet.Context, backend data.Backend) {
	ctx.Set(dataBackendKey, backend)
}

// GetDataBackend from the request context.
func GetDataBackend(ctx *juliet.Context) data.Backend {
	if backend, ok := ctx.Get(dataBackendKey); ok {
		return backend.(data.Backend)
	}
	return nil
}

func SetStreamBackend(ctx *juliet.Context, backend data.Backend) {
	ctx.Set(streamBackendKey, backend)
}

// GetStreamBackend from the request context.
func GetStreamBackend(ctx *juliet.Context) data.Backend {
	if backend, ok := ctx.Get(streamBackendKey); ok {
		return backend.(data.Backend)
	}
	return nil
}

func SetSourceIP(ctx *juliet.Context, sourceIP net.IP) {
	ctx.Set(sourceIPKey, sourceIP)
}

// GetSourceIP from the request context.
func GetSourceIP(ctx *juliet.Context) net.IP {
	if sourceIP, ok := ctx.Get(sourceIPKey); ok {
		return sourceIP.(net.IP)
	}
	return nil
}

func SetWhitelisted(ctx *juliet.Context, value bool) {
	ctx.Set(whitelistedKey, value)
}

// IsWhitelisted return true if the IP address in the request context is whitelisted.
func IsWhitelisted(ctx *juliet.Context) bool {
	if whitelisted, ok := ctx.Get(whitelistedKey); ok {
		return whitelisted.(bool)
	}

	uploadWhitelist := GetConfig(ctx).GetUploadWhitelist()

	// Check if the source IP address is in whitelist
	whitelisted := false
	if len(uploadWhitelist) > 0 {
		sourceIP := GetSourceIP(ctx)
		if sourceIP != nil {
			for _, subnet := range uploadWhitelist {
				if subnet.Contains(sourceIP) {
					whitelisted = true
					break
				}
			}
		}
	} else {
		whitelisted = true
	}
	ctx.Set(whitelistedKey, whitelisted)
	return whitelisted
}

func SetUser(ctx *juliet.Context, user *common.User) {
	ctx.Set(userKey, user)
}

// GetUser from the request context.
func GetUser(ctx *juliet.Context) *common.User {
	if user, ok := ctx.Get(userKey); ok {
		return user.(*common.User)
	}
	return nil
}

func SetToken(ctx *juliet.Context, token *common.Token) {
	ctx.Set(tokenKey, token)
}

// GetToken from the request context.
func GetToken(ctx *juliet.Context) *common.Token {
	if token, ok := ctx.Get(tokenKey); ok {
		return token.(*common.Token)
	}
	return nil
}

func SetFile(ctx *juliet.Context, file *common.File) {
	ctx.Set(fileKey, file)
}

// GetFile from the request context.
func GetFile(ctx *juliet.Context) *common.File {
	if file, ok := ctx.Get(fileKey); ok {
		return file.(*common.File)
	}
	return nil
}

func SetUpload(ctx *juliet.Context, upload *common.Upload) {
	ctx.Set(uploadKey, upload)
}

// GetUpload from the request context.
func GetUpload(ctx *juliet.Context) *common.Upload {
	if upload, ok := ctx.Get(uploadKey); ok {
		return upload.(*common.Upload)
	}
	return nil
}

func SetUploadAdmin(ctx *juliet.Context, value bool) {
	ctx.Set(uploadAdminKey, value)
}

// IsUploadAdmin returns true if the context has verified that current request can modify the upload
func IsUploadAdmin(ctx *juliet.Context) bool {
	if admin, ok := ctx.Get(uploadAdminKey); ok {
		return admin.(bool)
	}
	return false
}

func SetAdmin(ctx *juliet.Context, value bool) {
	ctx.Set(adminKey, value)
}

// IsAdmin check if the user is a Plik server administrator
func IsAdmin(ctx *juliet.Context) bool {
	if admin, ok := ctx.Get(adminKey); ok {
		return admin.(bool)
	}
	return false
}

func SetRedirectOnFailure(ctx *juliet.Context, value bool) {
	ctx.Set(redirectOnFailureKey, value)
}

// IsRedirectOnFailure return true if the http response should return
// a http redirect instead of an error string.
func IsRedirectOnFailure(ctx *juliet.Context) bool {
	if redirect, ok := ctx.Get(redirectOnFailureKey); ok {
		return redirect.(bool)
	}
	return false
}

func SetQuick(ctx *juliet.Context, value bool) {
	ctx.Set(quickKey, value)
}

// IsQuick changes the output of the addFile handler
func IsQuick(ctx *juliet.Context) bool {
	if quick, ok := ctx.Get(quickKey); ok {
		return quick.(bool)
	}
	return false
}

var userAgents = []string{"wget", "curl", "python-urllib", "libwwww-perl", "php", "pycurl", "Go-http-client"}

// Fail return write an error to the http response body.
// If IsRedirectOnFailure is true it write a http redirect that can be handled by the web client instead.
func Fail(ctx *juliet.Context, req *http.Request, resp http.ResponseWriter, message string, status int) {
	if IsRedirectOnFailure(ctx) {
		// The web client uses http redirect to get errors
		// from http redirect and display a nice HTML error message
		// But cli clients needs a clean string response
		userAgent := strings.ToLower(req.UserAgent())
		redirect := true
		for _, ua := range userAgents {
			if strings.HasPrefix(userAgent, ua) {
				redirect = false
			}
		}
		if redirect {
			config := GetConfig(ctx)
			http.Redirect(resp, req, fmt.Sprintf("%s/#/?err=%s&errcode=%d&uri=%s", config.Path, message, status, req.RequestURI), http.StatusMovedPermanently)
			return
		}
	}

	http.Error(resp, common.NewResult(message, nil).ToJSONString(), status)
}

// TestFail is a helper to test a httptest.ResponseRecoreder status
func TestFail(t *testing.T, resp *httptest.ResponseRecorder, status int, message string) {
	require.Equal(t, status, resp.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, err, 0, len(respBody), "empty response body")

	var result = &common.Result{}
	err = json.Unmarshal(respBody, result)
	require.NoError(t, err, "unable to unmarshal error")

	if message != "" {
		require.Contains(t, result.Message, message, "invalid response error message")
	}
}
