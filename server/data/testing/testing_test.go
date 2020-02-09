package file

import (
	"bytes"
	"errors"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/stretchr/testify/require"
)

// Ensure Testing Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

func TestAddFileError(t *testing.T) {
	backend := NewBackend()
	backend.SetError(errors.New("error"))

	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.Error(t, err, "missing error")
	require.Equal(t, "error", err.Error(), "invalid error message")
}

func TestAddFileReaderError(t *testing.T) {
	backend := NewBackend()

	upload := &common.Upload{}
	file := upload.NewFile()
	reader := common.NewErrorReader(errors.New("io error"))

	_, err := backend.AddFile(upload, file, reader)
	require.Error(t, err, "missing error")
	require.Equal(t, "io error", err.Error(), "invalid error message")
}

func TestAddFile(t *testing.T) {
	backend := NewBackend()
	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.NoError(t, err, "unable to add file")
}

func TestGetFileError(t *testing.T) {
	backend := NewBackend()
	backend.SetError(errors.New("error"))

	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.GetFile(upload, file.ID)
	require.Error(t, err, "missing error")
	require.Equal(t, "error", err.Error(), "invalid error message")
}

func TestGetFile(t *testing.T) {
	backend := NewBackend()
	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.NoError(t, err, "unable to add file")

	_, err = backend.GetFile(upload, file.ID)
	require.NoError(t, err, "unable to get file")
}

func TestRemoveFileError(t *testing.T) {
	backend := NewBackend()
	backend.SetError(errors.New("error"))

	upload := &common.Upload{}
	file := upload.NewFile()

	err := backend.RemoveFile(upload, file.ID)
	require.Error(t, err, "missing error")
	require.Equal(t, "error", err.Error(), "invalid error message")
}

func TestRemoveFile(t *testing.T) {
	backend := NewBackend()

	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.NoError(t, err, "unable to add file")

	_, err = backend.GetFile(upload, file.ID)
	require.NoError(t, err, "unable to get file")

	err = backend.RemoveFile(upload, file.ID)
	require.NoError(t, err, "unable to remove file")

	_, err = backend.GetFile(upload, file.ID)
	require.Error(t, err, "unable to get file")
	require.Equal(t, "file not found", err.Error(), "invalid error message")
}

func TestRemoveUploadError(t *testing.T) {
	backend := NewBackend()
	backend.SetError(errors.New("error"))

	upload := &common.Upload{}

	err := backend.RemoveUpload(upload)
	require.Error(t, err, "missing error")
	require.Equal(t, "error", err.Error(), "invalid error message")
}

func TestRemoveUpload(t *testing.T) {
	backend := NewBackend()

	upload := &common.Upload{}
	file := upload.NewFile()

	upload2 := common.NewUpload()
	file2 := upload2.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.NoError(t, err, "unable to add file")

	_, err = backend.AddFile(upload2, file2, &bytes.Buffer{})
	require.NoError(t, err, "unable to add file")

	_, err = backend.GetFile(upload, file.ID)
	require.NoError(t, err, "unable to get file")

	_, err = backend.GetFile(upload, file2.ID)
	require.NoError(t, err, "unable to get file")

	err = backend.RemoveUpload(upload)
	require.NoError(t, err, "unable to remove file")

	_, err = backend.GetFile(upload, file.ID)
	require.Error(t, err, "unable to get file")
	require.Equal(t, "file not found", err.Error(), "invalid error message")

	_, err = backend.GetFile(upload2, file2.ID)
	require.NoError(t, err, "unable to get file")
}
