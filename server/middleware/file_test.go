package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func TestFileNoUpload(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Internal error")
}

func TestFileNoFileID(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUpload(ctx, common.NewUpload())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Missing file id")
}

func TestFileNoFileName(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUpload(ctx, common.NewUpload())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"fileID": "fileID",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusBadRequest, "Missing file name")
}

func TestFileNotFound(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.SetUpload(ctx, common.NewUpload())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"fileID":   "fileID",
		"filename": "filename",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusNotFound, "File fileID not found")
}

func TestFileInvalidFileName(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	file := upload.NewFile()
	file.Name = "filename"

	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"fileID":   file.ID,
		"filename": "invalid_file_name",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusNotFound, "File invalid_file_name not found")
}

func TestFile(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	upload := common.NewUpload()
	file := upload.NewFile()
	file.Name = "filename"

	context.SetUpload(ctx, upload)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Fake gorilla/mux vars
	vars := map[string]string{
		"fileID":   file.ID,
		"filename": file.Name,
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	File(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	f := context.GetFile(ctx)
	require.NotNil(t, f, "missing file from context")
	require.Equal(t, file, f, "invalid file from context")
}
