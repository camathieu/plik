package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/root-gg/juliet"
	"github.com/root-gg/logger"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
)

func newTestingContext(config *common.Configuration) (ctx *juliet.Context) {
	ctx = juliet.NewContext()
	context.SetConfig(ctx, config)
	context.SetLogger(ctx, logger.NewLogger())
	context.SetMetadataBackend(ctx, metadata_test.NewBackend())
	context.SetDataBackend(ctx, data_test.NewBackend())
	context.SetStreamBackend(ctx, data_test.NewBackend())
	return ctx
}

func TestGetVersion(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/version", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetVersion(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	var result *common.BuildInfo
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err, "unable to unmarshal response body")

	require.EqualValues(t, common.GetBuildInfo(), result, "invalid build info")
}

func TestGetConfiguration(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/version", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetConfiguration(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	var result *common.Configuration
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err, "unable to unmarshal response body")
}

func TestGetQrCode(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/qrcode?url="+url.QueryEscape("https://root.gg"), bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetQrCode(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, 0, len(respBody), "invalid empty response body")
	require.Equal(t, "image/png", rr.Header().Get("Content-Type"), "invalid response content type")
}

func TestGetQrCodeWithSize(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/qrcode?url="+url.QueryEscape("https://root.gg")+"&size=100", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetQrCode(ctx, rr, req)

	// Check the status code is what we expect.
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.NotEqual(t, 0, len(respBody), "invalid empty response body")
	require.Equal(t, "image/png", rr.Header().Get("Content-Type"), "invalid response content type")
}

func TestGetQrCodeWithInvalidSize(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/qrcode?url="+url.QueryEscape("https://root.gg")+"&size=10000", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetQrCode(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "QRCode size must be lower than 1000")
}

func TestGetQrCodeWithInvalidSize2(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/qrcode?url="+url.QueryEscape("https://root.gg")+"&size=-1", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	GetQrCode(ctx, rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "QRCode size must be positive")
}
