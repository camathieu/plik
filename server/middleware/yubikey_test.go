package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestYubikeyNoUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestYubikeyNotEnabled(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetMetadataBackend(ctx).(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	upload := common.NewUpload()
	upload.Yubikey = "token"
	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "Yubikey are disabled on this server")
}

func TestYubikeyMissingToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetConfig(ctx).YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusUnauthorized, "Invalid yubikey token")
}

func TestYubikeyInvalidToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetConfig(ctx).YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"yubikey": "token",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusUnauthorized, "Invalid yubikey token")
}

func TestYubikeyInvalidDevice(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetConfig(ctx).YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"yubikey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusUnauthorized, "Invalid yubikey token")
}
