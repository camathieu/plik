package file

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"sync"

	"github.com/root-gg/plik/server/common"
)

// Backend object
type Backend struct {
	files map[string][]byte
	err   error
	mu    sync.Mutex
}

// NewBackend instantiate a new Testing Data Backend
// from configuration passed as argument
func NewBackend() (b *Backend) {
	b = new(Backend)
	b.files = make(map[string][]byte)
	return
}

// GetFile implementation for testing data backend will search
// on filesystem the asked file and return its reading filehandle
func (b *Backend) GetFile(upload *common.Upload, id string) (file io.ReadCloser, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	if data, ok := b.files[id]; ok {
		return ioutil.NopCloser(bytes.NewBuffer(data)), nil
	}

	return nil, errors.New("file not found")
}

// AddFile implementation for testing data backend will creates a new file for the given upload
// and save it on filesystem with the given file reader
func (b *Backend) AddFile(upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	if _, ok := b.files[file.ID]; ok {
		return nil, errors.New("file exists")
	}

	data, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}

	b.files[file.ID] = data

	return nil, nil
}

// RemoveFile implementation for testing data backend will delete the given
// file from filesystem
func (b *Backend) RemoveFile(upload *common.Upload, id string) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	delete(b.files, id)

	return nil
}

// RemoveUpload implementation for testing data backend will
// delete the whole upload. Given that an upload is a directory,
// we remove the whole directory at once.
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	for id := range upload.Files {
		delete(b.files, id)
	}

	return nil
}

// SetError set the error that this backend will return on any subsequent method call
func (b *Backend) SetError(err error) {
	b.err = err
}
