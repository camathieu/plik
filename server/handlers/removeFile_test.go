package handlers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestRemoveFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

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

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody))

	upload, err = context.GetMetadataBackend(ctx).GetUpload(upload.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(upload.Files), "invalid upload files count")
	require.Equal(t, common.FILE_DELETED, upload.Files[file1.ID].Status, "invalid removed file status")

	_, err = context.GetDataBackend(ctx).GetFile(upload, file1.ID)
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

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "You are not allowed to remove file from this upload")
}

func TestRemoveRemovedFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "removed"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusNotFound, "File file1 has already been removed")
}

func TestRemoveLastFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "uploaded"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/"+upload.ID+"/"+file1.ID+"/"+file1.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, "ok", string(respBody))

	u, err := context.GetMetadataBackend(ctx).GetUpload(upload.ID)
	require.NoError(t, err, "removed upload still exists")
	require.Nil(t, u, "removed upload still exists")

	_, err = context.GetDataBackend(ctx).GetFile(upload, file1.ID)
	require.Error(t, err, "removed file still exists")
}

func TestRemoveFileNoUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestRemoveFileNoFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

	upload := common.NewUpload()
	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestRemoveFileMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "uploaded"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	context.GetMetadataBackend(ctx).(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to update upload metadata")
}

func TestRemoveFileDataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUploadAdmin(ctx, true)

	data := "data"

	upload := common.NewUpload()
	upload.Create()

	file1 := upload.NewFile()
	file1.Name = "file1"
	file1.Status = "uploaded"

	err := createTestFile(ctx, upload, file1, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file 1")

	createTestUpload(ctx, upload)

	context.SetUpload(ctx, upload)
	context.SetFile(ctx, file1)

	req, err := http.NewRequest("DELETE", "/file/uploadID/fileID/fileName", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	context.GetDataBackend(ctx).(*data_test.Backend).SetError(errors.New("data backend error"))

	rr := httptest.NewRecorder()
	RemoveFile(ctx, rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")
}
