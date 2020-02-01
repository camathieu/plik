package plik

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

// File contains all relevant info needed to upload data to a Plik server
type File struct {
	Name string
	Size int64

	reader io.ReadCloser // Byte stream to upload
	error  error         // Upload error
	done   func()        // Upload callback

	upload  *Upload      // Link to upload and client
	details *common.File // File params returned by the server

	lock sync.Mutex
}

// NewFileFromReader creates a File from a filename and an io.ReadCloser
func newFileFromReadCloser(upload *Upload, name string, reader io.ReadCloser) *File {
	file := &File{}
	file.upload = upload
	file.Name = name
	file.reader = reader
	return file
}

// NewFileFromReader creates a File from a filename and an io.Reader
func newFileFromReader(upload *Upload, name string, reader io.Reader) *File {
	return newFileFromReadCloser(upload, name, ioutil.NopCloser(reader))
}

// NewFileFromPath creates a File from a filesystem path
func newFileFromPath(upload *Upload, path string) (file *File, err error) {

	// Test if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("File %s not found", path)
	}

	// Check mode
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("Unhandled file mode %s for file %s", fileInfo.Mode().String(), path)
	}

	// Open file
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open %s : %s", path, err)
	}

	filename := filepath.Base(path)
	file = newFileFromReader(upload, filename, fh)
	file.Size = fileInfo.Size()

	return file, err
}

// newFileFromParams create a new file object from the give file parameters
func newFileFromParams(upload *Upload, params *common.File) *File {
	file := &File{}
	file.upload = upload
	file.details = params
	file.Name = params.Name
	file.Size = params.CurrentSize
	return file
}

// Details return the file details returned by the server
func (file *File) Details() (details *common.File) {
	return file.details
}

// getParams return a common.File to be passed to internal methods
func (file *File) getParams() (params *common.File) {
	params = &common.File{}
	params.ID = file.ID()
	params.Name = file.Name
	return params
}

// ID return the file ID if any
func (file *File) ID() string {
	if file.details == nil {
		return ""
	}
	return file.details.ID
}

// ID return the file ID if any
func (file *File) Error() error {
	return file.error
}

// HasBeenUploaded return weather or not an attempt to upload the file ( successful or unsuccessful ) has been made
func (file *File) HasBeenUploaded() bool {
	if file.details != nil {
		if !(file.details.Status == "" || file.details.Status == "missing") {
			return true
		}
	}
	return false
}

// GetURL returns the URL to download the file
func (file *File) GetURL() (URL *url.URL, err error) {
	upload := file.upload

	if upload.ID() == "" {
		return nil, fmt.Errorf("Upload has not been created yet")
	}

	if file.ID() == "" {
		return nil, fmt.Errorf("File has not been uploaded yet")
	}

	mode := "file"
	if upload.Stream {
		mode = "stream"
	}

	var domain string
	if upload.details.DownloadDomain != "" {
		domain = upload.details.DownloadDomain
	} else {
		domain = upload.client.URL
	}

	fileURL := fmt.Sprintf("%s/%s/%s/%s/%s", domain, mode, upload.ID(), file.ID(), file.Name)

	// Parse to get a nice escaped url
	return url.Parse(fileURL)
}

// Upload uploads a single file.
func (file *File) Upload() (err error) {
	if file.HasBeenUploaded() {
		return fmt.Errorf("File has already been uploaded")
	}

	if !file.upload.HasBeenCreated() {
		err = file.upload.Create()
		if err != nil {
			return err
		}
	}

	defer file.reader.Close()

	fileInfo, err := file.upload.client.uploadFile(file.upload.getParams(), file.getParams(), file.reader)
	if err == nil {
		file.details = fileInfo
	} else {
		file.error = err
	}

	// Call the done callback before upload.Upload() returns
	if file.done != nil {
		file.done()
	}

	return err
}

// WrapReader a convenient function to alter the content of the file on the file ( encrypt / display progress / ... )
func (file *File) WrapReader(wrapper func(reader io.ReadCloser) io.ReadCloser) {
	file.reader = wrapper(file.reader)
}

// RegisterDoneCallback a callback to be executed after the file have been uploaded or failed ( check file.Error() )
func (file *File) RegisterDoneCallback(done func()) {
	file.done = done
}

// Download downloads all the upload files in a zip archive
func (file *File) Download() (reader io.ReadCloser, err error) {
	return file.upload.client.downloadFile(file.upload.getParams(), file.getParams())
}

// Delete remove the upload and all the associated files from the remote server
func (file *File) Delete() (err error) {
	return file.upload.client.removeFile(file.upload.getParams(), file.getParams())
}
