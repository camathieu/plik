package handlers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestRemoveUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "uploaded"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveUpload(ctx, rr, req)
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody), "invalid response body")

	u, err := ctx.GetMetadataBackend().GetUpload(upload.ID)
	require.NoError(t, err, "unexpected get upload error")
	require.Nil(t, u, "removed upload still exists")

	_, err = ctx.GetDataBackend().GetFile(upload, file1.ID)
	require.Error(t, err, "removed file still exists")
}

func TestRemoveUploadNotAdmin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "uploaded"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveUpload(ctx, rr, req)
	context.TestForbidden(t, rr, "you are not allowed to remove this upload")
}

func TestRemoveUploadNoUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("DELETE", "/upload/uploadID", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveUpload(ctx, rr, req)
	context.TestInternalServerError(t, rr, "missing upload from context")
}

func TestRemoveUploadMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	upload := common.NewUpload()
	upload.Create()

	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := ctx.NewRecorder(req)
	RemoveUpload(ctx, rr, req)
	context.TestInternalServerError(t, rr, "unable to remove upload metadata : metadata backend error")
}

func TestRemoveUploadDataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	upload := common.NewUpload()
	upload.Create()
	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	ctx.GetDataBackend().(*data_test.Backend).SetError(errors.New("data backend error"))

	rr := ctx.NewRecorder(req)
	RemoveUpload(ctx, rr, req)
	context.TestOK(t, rr)
}
