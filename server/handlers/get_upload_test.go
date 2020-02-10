package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func TestGetUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := &common.Upload{}
	upload.PrepareInsertForTests()
	upload.Login = "secret"
	upload.Password = "secret"
	file := upload.NewFile()
	file.Name = "file"
	createTestUpload(t, ctx, upload)
	ctx.SetUpload(upload)

	req, err := http.NewRequest("GET", "/upload/"+upload.ID, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	GetUpload(ctx, rr, req)

	context.TestOK(t, rr)

	respBody, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err, "unable to read response body")

	var uploadResult = &common.Upload{}
	err = json.Unmarshal(respBody, uploadResult)
	require.NoError(t, err, "unable to unmarshal response body")

	require.Equal(t, upload.ID, uploadResult.ID, "invalid upload id")
	require.NotZero(t, upload.CreatedAt, "missing creation date")
	require.Equal(t, upload.UploadToken, uploadResult.UploadToken, "invalid upload token")
	require.Equal(t, "", uploadResult.Login, "invalid upload login")
	require.Equal(t, "", uploadResult.Password, "invalid upload password")
	require.Len(t, uploadResult.Files, 1, "invalid upload files")
	require.Equal(t, file.ID, uploadResult.Files[0].ID, "invalid upload files")
	require.Equal(t, file.Name, uploadResult.Files[0].Name, "invalid upload files")
}

func TestGetUploadMissingUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "/upload/uploadID", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "unable to create new request")

	rr := ctx.NewRecorder(req)
	context.TestPanic(t, rr, "missing upload from context", func() {
		GetUpload(ctx, rr, req)
	})
}
