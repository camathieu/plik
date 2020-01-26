package file

import (
	"fmt"
	"io"
	"os"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/utils"
)

// Ensure File Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

// Config describes configuration for File Databackend
type Config struct {
	Directory string
}

// NewConfig instantiate a new default configuration
// and override it with configuration passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	config.Directory = "files" // Default upload directory is ./files
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	Config *Config
}

// NewBackend instantiate a new File Data Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend) {
	b = new(Backend)
	b.Config = config
	return
}

// GetFile implementation for file data backend will search
// on filesystem the asked file and return its reading filehandle
func (b *Backend) GetFile(upload *common.Upload, id string) (file io.ReadCloser, err error) {
	// Get upload directory
	directory, err := b.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get upload directory : %s", err)
	}

	// Get file path
	fullPath := directory + "/" + id

	// The file content will be piped directly
	// to the client response body
	file, err = os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s : %s", fullPath, err)
	}

	return file, nil
}

// AddFile implementation for file data backend will creates a new file for the given upload
// and save it on filesystem with the given file reader
func (b *Backend) AddFile(upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error) {
	// Get upload directory
	directory, err := b.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get upload directory : %s", err)
	}

	// Get file path
	fullPath := directory + "/" + file.ID

	// Create directory
	_, err = os.Stat(directory)
	if err != nil {
		err = os.MkdirAll(directory, 0777)
		if err != nil {
			return nil, fmt.Errorf("unable to create upload directory %s : %s", directory, err)
		}
	}

	// Create file
	out, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file %s : %s", fullPath, err)
	}

	// Copy file data from the client request body
	// to the file system
	_, err = io.Copy(out, fileReader)
	if err != nil {
		return nil, fmt.Errorf("unable to save file %s : %s", fullPath, err)
	}

	backendDetails = make(map[string]interface{})
	backendDetails["path"] = fullPath
	return backendDetails, nil
}

// RemoveFile implementation for file data backend will delete the given
// file from filesystem
func (b *Backend) RemoveFile(upload *common.Upload, id string) (err error) {
	// Get upload directory
	directory, err := b.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		return fmt.Errorf("unable to get upload directory : %s", err)
	}

	// Get file path
	fullPath := directory + "/" + id

	// Remove file
	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("unable to remove %s : %s", fullPath, err)
	}

	return nil
}

// RemoveUpload implementation for file data backend will
// delete the whole upload. Given that an upload is a directory,
// we remove the whole directory at once.
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	// Get upload directory
	fullPath, err := b.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		return fmt.Errorf("unable to get upload directory : %s", err)
	}

	// Remove everything at once
	err = os.RemoveAll(fullPath)
	if err != nil {
		return fmt.Errorf("unable to remove %s : %s", fullPath, err)
	}

	return nil
}

func (b *Backend) getDirectoryFromUploadID(uploadID string) (string, error) {
	// To avoid too many files in the same directory
	// data directory is splitted in two levels the
	// first level is the 2 first chars from the upload id
	// it gives 3844 possibilities reaching 65535 files per
	// directory at ~250.000.000 files uploaded.

	if len(uploadID) < 3 {
		return "", fmt.Errorf("invalid upload ID %s", uploadID)
	}
	return b.Config.Directory + "/" + uploadID[:2] + "/" + uploadID, nil
}
