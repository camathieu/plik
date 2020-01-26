package swift

import (
	"fmt"
	"io"

	"github.com/ncw/swift"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/utils"
)

// Ensure Swift Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

// Config describes configuration for Swift data backend
type Config struct {
	Username, Password, Host, ProjectName, Container string
}

// NewConfig instantiate a new default configuration
// and override it with configuration passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	config.Container = "plik"
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	config     *Config
	connection *swift.Connection
}

// NewBackend instantiate a new OpenSwift Data Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend) {
	b = new(Backend)
	b.config = config
	return b
}

// GetFile implementation for Swift Data Backend
func (b *Backend) GetFile(upload *common.Upload, fileID string) (reader io.ReadCloser, err error) {
	err = b.auth()
	if err != nil {
		return nil, err
	}

	reader, pipeWriter := io.Pipe()
	uuid := b.getFileID(upload, fileID)
	go func() {
		_, err = b.connection.ObjectGet(b.config.Container, uuid, pipeWriter, true, nil)
		defer func() { _ = pipeWriter.Close() }()
		if err != nil {
			return
		}
	}()

	return reader, nil
}

// AddFile implementation for Swift Data Backend
func (b *Backend) AddFile(upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error) {
	err = b.auth()
	if err != nil {
		return nil, err
	}

	uuid := b.getFileID(upload, file.ID)
	object, err := b.connection.ObjectCreate(b.config.Container, uuid, true, "", "", nil)

	_, err = io.Copy(object, fileReader)
	if err != nil {
		return nil, err
	}
	err = object.Close()
	if err != nil {
		return nil, err
	}

	return backendDetails, nil
}

// RemoveFile implementation for Swift Data Backend
func (b *Backend) RemoveFile(upload *common.Upload, fileID string) (err error) {
	err = b.auth()
	if err != nil {
		return err
	}

	uuid := b.getFileID(upload, fileID)
	err = b.connection.ObjectDelete(b.config.Container, uuid)
	if err != nil {
		return err
	}

	return
}

// RemoveUpload implementation for Swift Data Backend
// Iterates on each upload file and call removeFile
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	err = b.auth()
	if err != nil {
		return
	}

	for fileID := range upload.Files {
		uuid := b.getFileID(upload, fileID)
		err = b.connection.ObjectDelete(b.config.Container, uuid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) getFileID(upload *common.Upload, fileID string) string {
	return upload.ID + "." + fileID
}

func (b *Backend) auth() (err error) {
	if b.connection != nil && b.connection.Authenticated() {
		return
	}

	connection := &swift.Connection{
		UserName: b.config.Username,
		ApiKey:   b.config.Password,
		AuthUrl:  b.config.Host,
		Tenant:   b.config.ProjectName,
	}

	// Authenticate
	err = connection.Authenticate()
	if err != nil {
		return fmt.Errorf("unable to autenticate : %s", err)
	}
	b.connection = connection

	// Create container
	err = b.connection.ContainerCreate(b.config.Container, nil)
	if err != nil {
		return err
	}

	return nil
}
