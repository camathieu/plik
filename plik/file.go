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
	upload  *Upload      // Link to upload and client
	//done   func()        // Upload callback

	lock     sync.Mutex   // The two following fields need to be protected
	metadata *common.File // File metadata returned by the server

	done   chan struct{}
	err    error
	callback func(file *common.File, err error)
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
		return nil, fmt.Errorf("file %s not found", path)
	}

	// Check mode
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("unhandled file mode %s for file %s", fileInfo.Mode().String(), path)
	}

	// Open file
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s : %s", path, err)
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
	file.metadata = params
	file.Name = params.Name
	file.Size = params.CurrentSize
	return file
}

// Metadata return the file metadata returned by the server
func (file *File) Metadata() (details *common.File) {
	file.lock.Lock()
	defer file.lock.Unlock()

	return file.metadata
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
	file.lock.Lock()
	defer file.lock.Unlock()

	if file.metadata == nil {
		return ""
	}
	return file.metadata.ID
}

// ID return the file ID if any
func (file *File) Error() error {
	file.lock.Lock()
	defer file.lock.Unlock()

	return file.err
}

// HasBeenUploaded return weather or not an attempt to upload the file ( successful or unsuccessful ) has been made
func (file *File) HasBeenUploaded() bool {
	file.lock.Lock()
	defer file.lock.Unlock()

	return file.metadata != nil && file.metadata.Status == common.FileUploaded
}

// IsUploading return weather or not the file is currently being uploaded
func (file *File) IsUploading() bool {
	file.lock.Lock()
	defer file.lock.Unlock()

	return file.metadata != nil && file.metadata.Status == common.FileUploading
}

// GetURL returns the URL to download the file
func (file *File) GetURL() (URL *url.URL, err error) {
	upload := file.upload

	if upload.ID() == "" {
		return nil, fmt.Errorf("upload has not been created yet")
	}

	if file.ID() == "" {
		return nil, fmt.Errorf("file has not been uploaded yet")
	}

	mode := "file"
	if upload.Stream {
		mode = "stream"
	}

	var domain string
	if upload.Metadata().DownloadDomain != "" {
		domain = upload.Metadata().DownloadDomain
	} else {
		domain = upload.client.URL
	}

	fileURL := fmt.Sprintf("%s/%s/%s/%s/%s", domain, mode, upload.ID(), file.ID(), file.Name)

	// Parse to get a nice escaped url
	return url.Parse(fileURL)
}

func (file *File) ready() (done chan struct{}, err error) {
	file.lock.Lock()
	defer file.lock.Unlock()

	if file.done != nil {
		return file.done, nil
	}

	if file.metadata == nil {
		file.metadata = &common.File{Status: common.FileMissing}
	}

	if file.metadata.Status != common.FileMissing {
		return nil, fmt.Errorf("file %s is not ready to upload (%s)", file.Name, file.metadata.Status)
	}

	// Grab the lock by setting this channel
	file.done = make(chan struct{})
	file.err = nil

	file.metadata.Status = common.FileUploading

	return nil, nil
}

// Upload uploads a single file.
func (file *File) Upload() (err error) {
	err = file.upload.Create()
	if err != nil {
		return err
	}

	done, err := file.ready()
	if err != nil {
		return err
	}
	if done != nil {
		<- done
		file.lock.Lock()
		defer file.lock.Unlock()
		return file.err
	}

	defer func() { _ = file.reader.Close() }()

	fileInfo, err := file.upload.client.uploadFile(file.upload.getParams(), file.getParams(), file.reader)

	file.lock.Lock()
	if err == nil {
		file.metadata = fileInfo
	} else {
		file.err = err
	}
	file.lock.Unlock()

	// TODO release before or after callback ?
	close(file.done)

	// Call the done callback before upload.Upload() returns
	if file.callback != nil {
		file.callback(fileInfo, err)
	}

	return err
}

// WrapReader a convenient function to alter the content of the file on the file ( encrypt / display progress / ... )
func (file *File) WrapReader(wrapper func(reader io.ReadCloser) io.ReadCloser) {
	file.reader = wrapper(file.reader)
}

// RegisterDoneCallback a callback to be executed after the file have been uploaded or failed ( check file.Error() )
//func (file *File) RegisterDoneCallback(done func()) {
//	file.done = done
//}

// Download downloads all the upload files in a zip archive
func (file *File) Download() (reader io.ReadCloser, err error) {
	return file.upload.client.downloadFile(file.upload.getParams(), file.getParams())
}

// Delete remove the upload and all the associated files from the remote server
func (file *File) Delete() (err error) {
	return file.upload.client.removeFile(file.upload.getParams(), file.getParams())
}
