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

func (ctx *Context) InternalServerError(message string, err error) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	ctx.internalServerError(message, err)
}

func (ctx *Context) internalServerError(message string, err error) {

	if ctx.config != nil && ctx.config.Debug {
		// In DEBUG mode return the error message to the user
		if err != nil {
			message = fmt.Sprintf("%s : %s", message, err)
			err = nil // no need to log twice
		}
	} else {
		// In PROD mode return "internal server error" to the user
		message = internalServerError
	}

	ctx.panic = true
	ctx.fail(message, err, http.StatusInternalServerError)
}

func (ctx *Context) BadRequest(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	ctx.badRequest(message, params...)
}

func (ctx *Context) badRequest(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.fail(message, nil, http.StatusBadRequest)
}

func (ctx *Context) NotFound(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	ctx.notFound(message, params...)
}

func (ctx *Context) notFound(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.fail(message, nil, http.StatusNotFound)
}

func (ctx *Context) Forbidden(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	message = fmt.Sprintf(message, params...)
	ctx.fail(message, nil, http.StatusForbidden)
}

func (ctx *Context) Unauthorized(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	message = fmt.Sprintf(message, params...)
	ctx.fail(message, nil, http.StatusUnauthorized)
}

func (ctx *Context) MissingParameter(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	message = fmt.Sprintf(message, params...)
	ctx.badRequest(fmt.Sprintf("missing %s", message))
}

func (ctx *Context) InvalidParameter(message string, params ...interface{}) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	message = fmt.Sprintf(message, params...)
	ctx.badRequest(fmt.Sprintf("invalid %s", message))
}

var userAgents = []string{"wget", "curl", "python-urllib", "libwwww-perl", "php", "pycurl", "go-http-client"}

func (ctx *Context) Fail(message string, err error, status int) {
	ctx.mu.RLock()
	defer ctx.mu.RLock()

	ctx.fail(message, err, status)
}

func (ctx *Context) fail(message string, err error, status int) {

	logMessage := fmt.Sprintf("%s -- %d", message, status)
	if err != nil {
		logMessage = fmt.Sprintf("%s -- %v -- %d", message, err, status)
	}

	if ctx.logger != nil {
		if err != nil {
			ctx.logger.Warning(logMessage)
		}
	} else {
		log.Println(logMessage)
	}

	if ctx.req != nil && ctx.resp != nil {

		redirect := false
		if ctx.isRedirectOnFailure {
			// The web client uses http redirect to get errors
			// from http redirect and display a nice HTML error message
			// But cli clients needs a clean string response
			userAgent := strings.ToLower(ctx.req.UserAgent())
			redirect = true
			for _, ua := range userAgents {
				if strings.HasPrefix(userAgent, ua) {
					redirect = false
				}
			}
		}

		if redirect {
			url := fmt.Sprintf("%s/#/?err=%s&errcode=%d&uri=%s", ctx.config.Path, message, status, ctx.req.RequestURI)
			http.Redirect(ctx.resp, ctx.req, url, http.StatusMovedPermanently)
		} else {
			http.Error(ctx.resp, common.NewResult(message, nil).ToJSONString(), status)
		}
	}

	// This will be recovered by the HTTP server
	if ctx.panic {
		panic(logMessage)
	}
}

// NewRecorder create a new response recorder for testing
func (ctx *Context) NewRecorder(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ctx.SetReq(req)
	ctx.SetResp(rr)
	return rr
}

// TestMissingParameter is a helper to test a httptest.ResponseRecoreder status
func TestMissingParameter(t *testing.T, resp *httptest.ResponseRecorder, parameter string) {
	TestFail(t, resp, http.StatusBadRequest, fmt.Sprintf("missing %s", parameter))
}

// TestInvalidParameter is a helper to test a httptest.ResponseRecoreder status
func TestInvalidParameter(t *testing.T, resp *httptest.ResponseRecorder, parameter string) {
	TestFail(t, resp, http.StatusBadRequest, fmt.Sprintf("invalid %s", parameter))
}

// TestNotFound is a helper to test a httptest.ResponseRecoreder status
func TestNotFound(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusNotFound, message)
}

// TestForbidden is a helper to test a httptest.ResponseRecoreder status
func TestForbidden(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusForbidden, message)
}

// TestUnauthorized is a helper to test a httptest.ResponseRecoreder status
func TestUnauthorized(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusUnauthorized, message)
}

// TestBadRequest is a helper to test a httptest.ResponseRecoreder status
func TestBadRequest(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusBadRequest, message)
}

// TestInternalServerError is a helper to test a httptest.ResponseRecoreder status
func TestInternalServerError(t *testing.T, resp *httptest.ResponseRecorder) {
	TestFail(t, resp, http.StatusInternalServerError, internalServerError)
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

// TestFail is a helper to test a httptest.ResponseRecoreder status
func TestOK(t *testing.T, resp *httptest.ResponseRecorder) {
	require.Equal(t, http.StatusOK, resp.Code, "handler returned wrong status code")
}

// TestPanic is a helper to test a httptest.ResponseRecoreder status
func TestPanic(t *testing.T, resp *httptest.ResponseRecorder, message string, handler func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
		TestFail(t, resp, http.StatusInternalServerError, message)
	}()
	handler()
}
