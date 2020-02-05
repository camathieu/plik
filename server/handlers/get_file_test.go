package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"strconv"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func createTestFile(ctx *context.Context, upload *common.Upload, file *common.File, reader io.Reader) (err error) {
	dataBackend := ctx.GetDataBackend()
	_, err = dataBackend.AddFile(upload, file, reader)
	return err
}

func TestGetFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.SetUploadAdmin(true)

	data := "data"

	upload := common.NewUpload()
	file := upload.NewFile()
	file.Name = "file"
	file.Status = "uploaded"
	file.Md5 = "12345"
	file.Type = "type"
	file.CurrentSize = int64(len(data))
	createTestUpload(ctx, upload)

	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/"+upload.ID+"/"+file.ID+"/"+file.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	context.TestOK(t, rr)

	require.Equal(t, file.Type, rr.Header().Get("Content-Type"), "invalid response content type")
	require.Equal(t, strconv.Itoa(int(file.CurrentSize)), rr.Header().Get("Content-Length"), "invalid response content length")

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	require.Equal(t, data, string(respBody), "invalid file content")
}

func TestGetOneShotFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	upload.Create()
	upload.OneShot = true
	file := upload.NewFile()
	file.Name = "file"
	file.Status = "uploaded"
	createTestUpload(ctx, upload)

	data := "data"
	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/"+upload.ID+"/"+file.ID+"/"+file.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)

	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, data, string(respBody), "invalid file content")
}

func TestGetRemovedFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	file := upload.NewFile()
	file.Name = "file"
	file.Status = common.FileRemoved
	createTestUpload(ctx, upload)

	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/"+upload.ID+"/"+file.ID+"/"+file.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)

	context.TestNotFound(t, rr, fmt.Sprintf("file %s (%s) is not available : removed", file.Name, file.ID))
}

func TestGetDeletedFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	file := upload.NewFile()
	file.Name = "file"
	file.Status = common.FileDeleted
	createTestUpload(ctx, upload)

	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/"+upload.ID+"/"+file.ID+"/"+file.Name, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)

	context.TestNotFound(t, rr, fmt.Sprintf("file %s (%s) is not available : deleted", file.Name, file.ID))
}

func TestGetFileInvalidDownloadDomain(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)
	config.DownloadDomain = "http://download.domain"

	err := config.Initialize()
	require.NoError(t, err, "Unable to initialize config")

	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	require.Equal(t, 301, rr.Code, "handler returned wrong status code")
}

func TestGetFileMissingUpload(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	context.TestPanic(t, rr, "missing upload from context", func() {
		GetFile(ctx, rr, req)
	})
}

func TestGetFileMissingFile(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)
	ctx.SetUpload(common.NewUpload())

	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	context.TestPanic(t, rr, "missing file from context", func() {
		GetFile(ctx, rr, req)
	})
}

func TestGetHtmlFile(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	upload := common.NewUpload()
	upload.Create()

	file := upload.NewFile()
	file.Type = "html"
	file.Status = "uploaded"
	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	context.TestOK(t, rr)

	require.Equal(t, "text/plain", rr.Header().Get("Content-Type"), "invalid content type")
}

func TestGetFileNoType(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	upload := common.NewUpload()
	upload.Create()

	file := upload.NewFile()
	file.Status = "uploaded"
	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	context.TestOK(t, rr)

	require.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"), "invalid content type")
}

func TestGetFileDataBackendError(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	upload := common.NewUpload()
	upload.Create()

	file := upload.NewFile()
	file.Name = "file"
	file.Status = "uploaded"
	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	ctx.GetDataBackend().(*data_test.Backend).SetError(errors.New("data backend error"))
	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	context.TestInternalServerError(t, rr, "unable to get file from data backend : data backend error")
}

func TestGetFileMetadataBackendError(t *testing.T) {
	config := common.NewConfiguration()
	ctx := newTestingContext(config)

	upload := common.NewUpload()
	upload.OneShot = true
	upload.Create()

	file := upload.NewFile()
	file.Status = "uploaded"
	err := createTestFile(ctx, upload, file, bytes.NewBuffer([]byte("data")))
	require.NoError(t, err, "unable to create test file")

	ctx.SetUpload(upload)
	ctx.SetFile(file)

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))
	req, err := http.NewRequest("GET", "/file/", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetFile(ctx, rr, req)
	context.TestInternalServerError(t, rr, "unable to update file metadata : metadata backend error")
}
