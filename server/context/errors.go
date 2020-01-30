package context

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

var internalServerError = "internal server error"

func (ctx *Context) MissingParameter(parameter string) {
	ctx.BadRequest(fmt.Sprintf("missing %s", parameter), nil)
}

func (ctx *Context) InvalidParameterValue(value interface{}, parameter string, limit string) {
	ctx.BadRequest(fmt.Sprintf("invalid %s value %v. %s", value, parameter, limit), nil)
}

func (ctx *Context) UploadIDNotFound(uploadID string) {
	ctx.NotFound(fmt.Sprintf("upload %s not found", uploadID), nil)
}

func (ctx *Context) UploadNotFound() {
	if ctx.upload == nil {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing upload from context"))
		return
	}

	ctx.NotFound(fmt.Sprintf("upload %s not found", ctx.upload.ID), nil)
}

func (ctx *Context) FileNotFound() {
	if ctx.upload == nil {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing upload from context"))
		return
	}
	if ctx.file == nil {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing file from context"))
		return
	}

	ctx.NotFound(fmt.Sprintf("file %s (%s) not found", ctx.file.Name, ctx.file.ID), nil)
}

func (ctx *Context) UserNotFound(user *common.User) {
	if ctx.upload == nil {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing user from context"))
		return
	}

	ctx.NotFound(fmt.Sprintf("user %s not found", ctx.user.Name), nil)
}

func (ctx *Context) InternalServerError(message string, err error) {
	ctx.Fail(message, err, http.StatusInternalServerError)
}

func (ctx *Context) BadRequest(message string, err error) {
	ctx.Fail(message, err, http.StatusBadRequest)
}

func (ctx *Context) NotFound(message string, err error) {
	ctx.Fail(message, err, http.StatusInternalServerError)
}

var userAgents = []string{"wget", "curl", "python-urllib", "libwwww-perl", "php", "pycurl", "go-http-client"}

func (ctx *Context) Fail(message string, err error, status int) {
	msg := fmt.Sprintf("%s : %v (%d)", message, err, status)

	if err != nil {
		if ctx.HasLogger() {
			ctx.logger.Warningf(msg)
		} else {
			log.Println(msg)
		}
	}

	if ctx.HasRequest() && ctx.HasRequest() {
		if ctx.isRedirectOnFailure {
			// The web client uses http redirect to get errors
			// from http redirect and display a nice HTML error message
			// But cli clients needs a clean string response
			userAgent := strings.ToLower(ctx.req.UserAgent())
			redirect := true
			for _, ua := range userAgents {
				if strings.HasPrefix(userAgent, ua) {
					redirect = false
				}
			}
			if redirect {
				url := fmt.Sprintf("%s/#/?err=%s&errcode=%d&uri=%s", ctx.config.Path, message, status, ctx.req.RequestURI)
				http.Redirect(ctx.resp, ctx.req, url, http.StatusMovedPermanently)
			}
		} else {
			http.Error(ctx.resp, common.NewResult(message, nil).ToJSONString(), status)
		}
	}

	if ctx.isPanic {
		panic(msg)
	}
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
