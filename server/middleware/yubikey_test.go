package middleware

import (
	"bytes"
	"errors"
	"net/http"
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

	rr := ctx.NewRecorder(req)

	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)
	context.TestInternalServerError(t, rr, "missing upload in yubikey middleware")
}

func TestYubikeyNoYubikey(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	ctx.SetUpload(common.NewUpload())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)

	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)
	context.TestOK(t, rr)
}

func TestYubikeyNotEnabled(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	upload := common.NewUpload()
	upload.Yubikey = "token"
	ctx.SetUpload(upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestBadRequest(t, rr, "yubikey are disabled on this server")
}

func TestYubikeyMissingToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	ctx.SetUpload(upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestBadRequest(t, rr, "missing yubikey token")
}

func TestYubikeyInvalidToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	ctx.SetUpload(upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"yubikey": "token",
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestBadRequest(t, rr, "invalid yubikey token")
}

func TestYubikeyInvalidDevice(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().YubikeyEnabled = true

	upload := common.NewUpload()
	upload.Yubikey = "token"
	ctx.SetUpload(upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"yubikey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Yubikey(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestBadRequest(t, rr, "invalid yubikey token")
}
