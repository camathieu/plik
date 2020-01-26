package stream

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/utils"
)

// Ensure Stream Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

// Config describes configuration for stream data backend
type Config struct {
}

// NewConfig instantiate a new default configuration
// and override it with configuration passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	Config *Config
	store  map[string]io.ReadCloser
	mu     sync.Mutex
}

// NewBackend instantiate a new Stream Data Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend) {
	b = new(Backend)
	b.Config = config
	b.store = make(map[string]io.ReadCloser)
	return
}

// GetFile implementation for steam data backend will search
// on filesystem the requested steam and return its reading filehandle
func (b *Backend) GetFile(upload *common.Upload, fileID string) (stream io.ReadCloser, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	storeID := upload.ID + "/" + fileID
	stream, ok := b.store[storeID]
	if !ok {
		return nil, fmt.Errorf("missing reader")
	}

	delete(b.store, fileID)

	return stream, err
}

// AddFile implementation for steam data backend will creates a new steam for the given upload
// and save it on filesystem with the given steam reader
func (b *Backend) AddFile(upload *common.Upload, file *common.File, stream io.Reader) (backendDetails map[string]interface{}, err error) {
	backendDetails = make(map[string]interface{})

	id := upload.ID + "/" + file.ID

	pipeReader, pipeWriter := io.Pipe()

	b.mu.Lock()

	b.store[id] = pipeReader
	defer delete(b.store, id)

	b.mu.Unlock()

	// This will block until download begins
	_, err = io.Copy(pipeWriter, stream)
	pipeWriter.Close()

	return backendDetails, nil
}

// RemoveFile is not implemented
func (b *Backend) RemoveFile(upload *common.Upload, id string) (err error) {
	return errors.New("can't remove stream file")
}

// RemoveUpload is not implemented
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	return errors.New("can't remove stream upload")
}
