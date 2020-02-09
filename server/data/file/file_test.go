package file

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func newBackend(t *testing.T) (backend *Backend, cleanup func()) {
	dir, err := ioutil.TempDir("", "pliktest")
	require.NoError(t, err, "unable to create temp directory")

	backend = NewBackend(&Config{Directory: dir})
	cleanup = func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err)
		}
	}

	return backend, cleanup
}

func TestNewFileBackendConfig(t *testing.T) {
	config := NewConfig(make(map[string]interface{}))
	require.NotNil(t, config, "invalid nil config")
}

func TestAddFileInvalidUploadId(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.Error(t, err, "no error with invalid upload id")
}

func TestAddFileImpossibleToCreateDirectory(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	// null byte looks like a good invalid dirname value ^^
	backend.Config.Directory = string([]byte{0})

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	_, err := backend.AddFile(upload, file, &bytes.Buffer{})
	require.Error(t, err, "unable to create directory")
}

func TestAddFileInvalidReader(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	reader := common.NewErrorReader(errors.New("io error"))
	_, err := backend.AddFile(upload, file, reader)
	require.Error(t, err, "unable to create directory")
	require.Contains(t, err.Error(), "io error", "invalid error")
}

func TestAddFile(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	reader := bytes.NewBufferString("data")
	path, err := backend.AddFile(upload, file, reader)
	require.NoError(t, err, "unable to add file")
	require.NotZero(t, path, "missing path")

	fh, err := os.Open(path)
	require.NoError(t, err, "unable to open file")
	defer fh.Close()

	read, err := ioutil.ReadAll(fh)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, "data", string(read), "inavlid file content")
}

func TestGetFileInvalidDirectory(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	// null byte looks like a good invalid dirname value ^^
	backend.Config.Directory = string([]byte{0})

	_, err := backend.GetFile(upload, file)
	require.Error(t, err, "no error with invalid upload directory")
}

func TestGetFileMissingFile(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	_, err := backend.GetFile(upload, file)
	require.Error(t, err, "no error with missing file")
	require.Contains(t, err.Error(), "no such file or directory", "invalid error message")
}

func TestGetFile(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	reader := bytes.NewBufferString("data")
	_, err := backend.AddFile(upload, file, reader)
	require.NoError(t, err, "unable to add file")

	fileReader, err := backend.GetFile(upload, file)
	require.NoError(t, err, "unable to get file")

	read, err := ioutil.ReadAll(fileReader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, "data", string(read), "inavlid file content")
}

func TestRemoveFileInvalidDirectory(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	// null byte looks like a good invalid dirname value ^^
	backend.Config.Directory = string([]byte{0})

	err := backend.RemoveFile(upload, file)
	require.Error(t, err, "no error with invalid upload id")
}

func TestRemoveFileMissingFile(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	err := backend.RemoveFile(upload, file)
	require.Error(t, err, "no error with invalid upload id")
	require.Contains(t, err.Error(), "no such file or directory", "invalid error message")
}

func TestRemoveFile(t *testing.T) {
	backend, clean := newBackend(t)
	defer clean()

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	reader := bytes.NewBufferString("data")
	path, err := backend.AddFile(upload, file, reader)
	require.NoError(t, err, "unable to add file")
	require.NotNil(t, path, "missing backend detail")

	fh, err := os.Open(path)
	require.NoError(t, err, "unable to open file")

	read, err := ioutil.ReadAll(fh)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, "data", string(read), "inavlid file content")

	err = backend.RemoveFile(upload, file)
	require.NoError(t, err, "unable to remove file")

	_, err = os.Open(path)
	require.Error(t, err, "able to open removed file")
}
