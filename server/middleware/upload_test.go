package middleware

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/root-gg/utils"
	"github.com/stretchr/testify/require"
)

func TestUploadNoUploadID(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestMissingParameter(t, rr, "upload id")
}

func TestUploadMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": "uploadID",
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)

	context.TestPanic(t, rr, "unable to get upload metadata : metadata backend error", func() {
		Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)
	})
}

func TestUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, upload, ctx.GetUpload(), "invalid upload from context")
}

func TestUploadExpired(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	upload.Create()
	upload.TTL = 60
	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestNotFound(t, rr, "upload "+upload.ID+" has expired")
}

func TestUploadToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	upload.Create()
	upload.UploadToken = "token"

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	req.Header.Set("X-UploadToken", upload.UploadToken)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, upload, ctx.GetUpload(), "invalid upload from context")
	require.True(t, ctx.IsUploadAdmin(), "invalid upload admin status")
}

func TestUploadUser(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	user := common.NewUser()
	user.ID = "user"
	ctx.SetUser(user)

	upload := common.NewUpload()
	upload.Create()
	upload.User = user.ID

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, upload, ctx.GetUpload(), "invalid upload from context")
	require.True(t, ctx.IsUploadAdmin(), "invalid upload admin status")
}

func TestUploadUserAdmin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	user := common.NewUser()
	user.ID = "user"
	ctx.SetAdmin(true)
	ctx.SetUser(user)

	upload := common.NewUpload()
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, upload, ctx.GetUpload(), "invalid upload from context")
	require.True(t, ctx.IsUploadAdmin(), "invalid upload admin status")
}

func TestUploadPasswordMissingHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	upload := common.NewUpload()
	upload.ProtectedByPassword = true
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestUnauthorized(t, rr, "please provide valid credentials to access this upload")
}

func TestUploadPasswordInvalidHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	upload := common.NewUpload()
	upload.ProtectedByPassword = true
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	req.Header.Set("Authorization", "invalid_header")

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestUnauthorized(t, rr, "please provide valid credentials to access this upload")
}

func TestUploadPasswordInvalidScheme(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	upload := common.NewUpload()
	upload.ProtectedByPassword = true
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	req.Header.Set("Authorization", "invalid_scheme invalid_creds")

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestUnauthorized(t, rr, "please provide valid credentials to access this upload")
}

func TestUploadPasswordInvalidPassword(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	upload := common.NewUpload()
	upload.ProtectedByPassword = true
	upload.Create()

	err := ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	req.Header.Set("Authorization", "Basic invalid_creds")

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestUnauthorized(t, rr, "please provide valid credentials to access this upload")
}

func TestUploadPassword(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	var err error

	upload := common.NewUpload()
	upload.ProtectedByPassword = true
	upload.Login = "login"
	upload.Password = "password"
	upload.Create()

	// The Authorization header will contain the base64 version of "login:password"
	// Save only the md5sum of this string to authenticate further requests
	b64str := base64.StdEncoding.EncodeToString([]byte(upload.Login + ":" + upload.Password))
	upload.Password, err = utils.Md5sum(b64str)
	require.NoError(t, err, "unable to b64encode upload credentials")

	err = ctx.GetMetadataBackend().CreateUpload(upload)
	require.NoError(t, err, "Unable to create upload")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"uploadID": upload.ID,
	}
	req = mux.SetURLVars(req, vars)

	req.Header.Add("Authorization", "Basic "+b64str)

	rr := ctx.NewRecorder(req)
	Upload(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, upload, ctx.GetUpload(), "invalid upload from context")
	require.False(t, ctx.IsUploadAdmin(), "invalid upload admin status")
}
