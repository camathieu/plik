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
func (b *Backend) GetFile(file *common.File) (reader io.ReadCloser, err error) {
	err = b.auth()
	if err != nil {
		return nil, err
	}

	reader, pipeWriter := io.Pipe()
	objectID := objectID(file)
	go func() {
		_, err = b.connection.ObjectGet(b.config.Container, objectID, pipeWriter, true, nil)
		defer func() { _ = pipeWriter.Close() }()
		if err != nil {
			return
		}
	}()

	return reader, nil
}

// AddFile implementation for Swift Data Backend
func (b *Backend) AddFile(file *common.File, fileReader io.Reader) (backendDetails string, err error) {
	err = b.auth()
	if err != nil {
		return "", err
	}

	objectId := objectID(file)
	object, err := b.connection.ObjectCreate(b.config.Container, objectId, true, "", "", nil)

	_, err = io.Copy(object, fileReader)
	if err != nil {
		return "", err
	}
	err = object.Close()
	if err != nil {
		return "", err
	}

	return backendDetails, nil
}

// RemoveFile implementation for Swift Data Backend
func (b *Backend) RemoveFile(file *common.File) (err error) {
	err = b.auth()
	if err != nil {
		return err
	}

	objectID := objectID(file)
	err = b.connection.ObjectDelete(b.config.Container, objectID)
	if err != nil {
		return err
	}

	return
}

func objectID(file *common.File) string {
	return file.UploadID + "." + file.ID
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
