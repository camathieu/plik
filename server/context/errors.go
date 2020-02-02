package context

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

var internalServerError = "internal server error"

// InternalServerError is a helper to generate http.StatusInternalServerError responses
func (ctx *Context) InternalServerError(message string, err error) {
	config := ctx.GetConfig()

	if config != nil && config.Debug {
		// In DEBUG mode return the error message to the user
		if err != nil {
			message = fmt.Sprintf("%s : %s", message, err)
			err = nil // no need to log twice
		}
	} else {
		// In PROD mode return "internal server error" to the user
		message = internalServerError
	}

	ctx.Fail(message, err, http.StatusInternalServerError)

	if config != nil && config.Debug {
		debug.PrintStack()
	}
}

// BadRequest is a helper to generate http.BadRequest responses
func (ctx *Context) BadRequest(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.Fail(message, nil, http.StatusBadRequest)
}

// NotFound is a helper to generate http.NotFound responses
func (ctx *Context) NotFound(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.Fail(message, nil, http.StatusNotFound)
}

// Forbidden is a helper to generate http.Forbidden responses
func (ctx *Context) Forbidden(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.Fail(message, nil, http.StatusForbidden)
}

// Unauthorized is a helper to generate http.Unauthorized responses
func (ctx *Context) Unauthorized(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.Fail(message, nil, http.StatusUnauthorized)
}

// MissingParameter is a helper to generate http.BadRequest responses
func (ctx *Context) MissingParameter(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.BadRequest(fmt.Sprintf("missing %s", message))
}

// InvalidParameter is a helper to generate http.BadRequest responses
func (ctx *Context) InvalidParameter(message string, params ...interface{}) {
	message = fmt.Sprintf(message, params...)
	ctx.BadRequest(fmt.Sprintf("invalid %s", message))
}

var userAgents = []string{"wget", "curl", "python-urllib", "libwwww-perl", "php", "pycurl", "go-http-client"}

// Fail is a helper to generate http error responses
func (ctx *Context) Fail(message string, err error, status int) {

	// Snapshot all we need
	ctx.mu.Lock()
	logger := ctx.logger
	config := ctx.config
	isRedirectOnFailure := ctx.isRedirectOnFailure
	req := ctx.req
	resp := ctx.resp
	ctx.mu.Unlock()

	// Generate log message
	logMessage := fmt.Sprintf("%s -- %d", message, status)
	if err != nil {
		logMessage = fmt.Sprintf("%s -- %v -- %d", message, err, status)
	}

	// Log message
	if logger != nil {
		if err != nil {
			logger.Warning(logMessage)
		}
	} else {
		log.Println(logMessage)
	}

	if req != nil && resp != nil {
		redirect := false
		if isRedirectOnFailure {
			// The web client uses http redirect to get errors
			// from http redirect and display a nice HTML error message
			// But cli clients needs a clean string response
			userAgent := strings.ToLower(req.UserAgent())
			redirect = true
			for _, ua := range userAgents {
				if strings.HasPrefix(userAgent, ua) {
					redirect = false
				}
			}
		}

		if config != nil && redirect {
			url := fmt.Sprintf("%s/#/?err=%s&errcode=%d&uri=%s", config.Path, message, status, req.RequestURI)
			http.Redirect(resp, req, url, http.StatusMovedPermanently)
		} else {
			http.Error(resp, common.NewResult(message, nil).ToJSONString(), status)
		}
	}
}

// NewRecorder create a new response recorder for testing
func (ctx *Context) NewRecorder(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ctx.SetReq(req)
	ctx.SetResp(rr)
	return rr
}

// TestMissingParameter is a helper to test a httptest.ResponseRecorder status
func TestMissingParameter(t *testing.T, resp *httptest.ResponseRecorder, parameter string) {
	TestFail(t, resp, http.StatusBadRequest, fmt.Sprintf("missing %s", parameter))
}

// TestInvalidParameter is a helper to test a httptest.ResponseRecorder status
func TestInvalidParameter(t *testing.T, resp *httptest.ResponseRecorder, parameter string) {
	TestFail(t, resp, http.StatusBadRequest, fmt.Sprintf("invalid %s", parameter))
}

// TestNotFound is a helper to test a httptest.ResponseRecorder status
func TestNotFound(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusNotFound, message)
}

// TestForbidden is a helper to test a httptest.ResponseRecorder status
func TestForbidden(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusForbidden, message)
}

// TestUnauthorized is a helper to test a httptest.ResponseRecorder status
func TestUnauthorized(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusUnauthorized, message)
}

// TestBadRequest is a helper to test a httptest.ResponseRecorder status
func TestBadRequest(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusBadRequest, message)
}

// TestInternalServerError is a helper to test a httptest.ResponseRecorder status
func TestInternalServerError(t *testing.T, resp *httptest.ResponseRecorder, message string) {
	TestFail(t, resp, http.StatusInternalServerError, message)
}

// TestFail is a helper to test a httptest.ResponseRecorder status
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

// TestOK is a helper to test a httptest.ResponseRecorder status
func TestOK(t *testing.T, resp *httptest.ResponseRecorder) {
	require.Equal(t, http.StatusOK, resp.Code, "handler returned wrong status code")
}

// TestPanic is a helper to test a httptest.ResponseRecorder status
func TestPanic(t *testing.T, resp *httptest.ResponseRecorder, message string, handler func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()
	handler()
}
