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

func TestRemoveFile(t *testing.T) {
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

	file2 := upload.NewFile()
	file2.Name = "file2"
	file2.Status = "uploaded"

	err = createTestFile(ctx, upload, file2, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 2")

	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody))

	upload, err = ctx.GetMetadataBackend().GetUpload(upload.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(upload.Files), "invalid upload files count")
	require.Equal(t, common.FileDeleted, upload.Files[file1.ID].Status, "invalid removed file status")

	_, err = ctx.GetDataBackend().GetFile(upload, file1.ID)
	require.Error(t, err, "removed file still exists")
}

func TestRemoveFileNotAdmin(t *testing.T) {
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
	RemoveFile(ctx, rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "You are not allowed to remove file from this upload")
}

func TestRemoveRemovedFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "removed"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	ctx.SetUpload(upload)
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusNotFound, "File file1 has already been removed")
}

func TestRemoveLastFile(t *testing.T) {
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
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody))

	u, err := ctx.GetMetadataBackend().GetUpload(upload.ID)
	require.NoError(t, err, "removed upload still exists")
	require.Nil(t, u, "removed upload still exists")

	_, err = ctx.GetDataBackend().GetFile(upload, file1.ID)
	require.Error(t, err, "removed file still exists")
}

func TestRemoveFileNoUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestRemoveFileNoFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	upload := common.NewUpload()
	ctx.SetUpload(upload)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestRemoveFileMetadataBackendError(t *testing.T) {
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
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to update upload metadata")
}

func TestRemoveFileDataBackendError(t *testing.T) {
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
	ctx.SetFile(file1)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	ctx.GetDataBackend().(*data_test.Backend).SetError(errors.New("data backend error"))

	rr := ctx.NewRecorder(req)
	RemoveFile(ctx, rr, req)
	context.TestOK(t, rr)
}
