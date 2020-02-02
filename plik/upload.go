package plik

import (
	"fmt"
	"io"
	"net/url"
	"strconv"
	"sync"

	"github.com/root-gg/plik/server/common"
)

// UploadParams store the different options available when uploading file to a Plik server
// One should add files to the upload before calling Create or Upload
type UploadParams struct {
	Stream    bool // Don't store the file on the server
	OneShot   bool // Force deletion of the file from the server after the first download
	Removable bool // Allow upload and upload files to be removed from the server at any time

	TTL      int    // Time in second before automatic deletion of the file from the server
	Comments string // Arbitrary comment to attach to the upload ( the web interface support markdown language )

	Token string // Authentication token to link an upload to a Plik user

	Login    string // HttpBasic protection for the upload
	Password string // Login and Password

	Yubikey string // Yubikey OTP
}

// Upload store the necessary data to upload files to a Plik server
type Upload struct {
	UploadParams
	client  *Client        // Client that makes the actual HTTP calls
	files   []*File        // Files to upload

	lock     sync.Mutex     // The following fields need to be protected
	metadata *common.Upload // Upload metadata ( once created )

	done     chan struct{}
	err      error
}

// newUpload create and initialize a new Upload object
func newUpload(client *Client) (upload *Upload) {
	upload = new(Upload)
	upload.client = client

	// Copy the default upload params from the client
	upload.UploadParams = *client.UploadParams

	return upload
}

func newUploadFromParams(client *Client, params *common.Upload) (upload *Upload) {
	upload = newUpload(client)
	upload.Stream = params.Stream
	upload.OneShot = params.OneShot
	upload.Removable = params.Removable
	upload.TTL = params.TTL
	upload.Comments = params.Comments
	upload.metadata = params

	for _, file := range params.Files {
		upload.add(newFileFromParams(upload, file))
	}

	return upload
}

// AddFiles add one or several files to be uploaded
func (upload *Upload) add(file *File) {
	upload.lock.Lock()
	defer upload.lock.Unlock()

	upload.files = append(upload.files, file)
}

// AddFileFromPath add a new file from a filesystem path
func (upload *Upload) AddFileFromPath(name string) (file *File, err error) {
	file, err = newFileFromPath(upload, name)
	if err != nil {
		return nil, err
	}
	upload.add(file)
	return file, nil
}

// AddFileFromReader add a new file from a filename and io.Reader
func (upload *Upload) AddFileFromReader(name string, reader io.Reader) (file *File) {
	file = newFileFromReader(upload, name, reader)
	upload.add(file)
	return file
}

// AddFileFromReadCloser add a new file from a filename and io.ReadCloser
func (upload *Upload) AddFileFromReadCloser(name string, reader io.ReadCloser) (file *File) {
	file = newFileFromReadCloser(upload, name, reader)
	upload.add(file)
	return file
}

// Metadata return the upload metadata returned by the server
func (upload *Upload) Metadata() (details *common.Upload) {
	upload.lock.Lock()
	defer upload.lock.Unlock()

	return upload.metadata
}

// getParams returns a common.Upload to be passed to internal methods
func (upload *Upload) getParams() (params *common.Upload) {
	params = common.NewUpload()
	params.Stream = upload.Stream
	params.OneShot = upload.OneShot
	params.Removable = upload.Removable
	params.TTL = upload.TTL
	params.Comments = upload.Comments
	params.Token = upload.Token
	params.Login = upload.Login
	params.Password = upload.Password

	metadata := upload.Metadata()
	if upload.HasBeenCreated() {
		params.ID = metadata.ID
		params.UploadToken = metadata.UploadToken
	}

	for i, file := range upload.Files() {
		fileParams := file.getParams()
		if fileParams.ID == "" {
			reference := strconv.Itoa(i)
			fileParams.Reference = reference
			params.Files[reference] = fileParams
		} else {
			params.Files[fileParams.ID] = fileParams
		}
	}
	return params
}

// Files Return the upload files
func (upload *Upload) Files() (files []*File) {
	upload.lock.Lock()
	defer upload.lock.Unlock()

	return upload.files
}

// HasBeenCreated return true if the upload has been created server side ( has an ID )
func (upload *Upload) HasBeenCreated() bool {
	return upload.Metadata() != nil
}

// ID returns the upload ID if the upload has been created server side
func (upload *Upload) ID() string {
	metadata := upload.Metadata()
	if metadata == nil {
		return ""
	}
	return metadata.ID
}

func (upload *Upload) ready() (done chan struct{}) {
	upload.lock.Lock()
	defer upload.lock.Unlock()

	if upload.done != nil {
		return upload.done
	}

	// Grab the lock by setting this channel
	upload.done = make(chan struct{})

	return nil
}

// Create a new empty upload on a Plik Server
func (upload *Upload) Create() (err error) {

	done := upload.ready()
	if done != nil {
		<- done
		upload.lock.Lock()
		defer upload.lock.Unlock()
		return upload.err
	}

	uploadParams := upload.getParams()

	// Crate the upload on the Plik server
	uploadMetadata, err := upload.client.create(uploadParams)
	if err == nil {
		// Keep all the uploadMetadata but we are mostly interested in the upload ID
		upload.lock.Lock()
		upload.metadata = uploadMetadata
		err = upload.updateMetadata(uploadMetadata)
		upload.lock.Unlock()
	}

	upload.lock.Lock()
	upload.err = err
	upload.lock.Unlock()

	close(upload.done)

	return err
}

func (upload *Upload) updateMetadata(uploadMetadata *common.Upload) (err error){
	// Here also we keep all the file info but we are also mostly interested in the file ID
	// We use the reference system to avoid problems if uploading several files with the same filename
LOOP:
	for _, file := range uploadMetadata.Files {
		for i, f := range upload.files {
			reference := strconv.Itoa(i)

			if file.Reference == reference {
				f.lock.Lock()
				f.metadata = file // Update the file metadata
				f.lock.Unlock()
				continue LOOP
			}
		}
		return fmt.Errorf("no file match for file reference %s", file.Reference)
	}

	return nil
}

// Upload uploads all files of the upload in parallel
func (upload *Upload) Upload() (err error) {
	err = upload.Create()
	if err != nil {
		return err
	}

	files := upload.Files()
	errors := make(chan error, len(files))

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file *File) {
			defer wg.Done()
			err := file.Upload()
			if err != nil {
				errors <- err
				return
			}
		}(file)
	}

	wg.Wait()

	close(errors)
	for err := range errors {
		fmt.Println(err)
		return fmt.Errorf("failed to upload at least one file. Check each file status for more details")
	}

	return nil
}

// GetURL returns the URL page of the upload
func (upload *Upload) GetURL() (u *url.URL, err error) {
	if !upload.HasBeenCreated() {
		return nil, fmt.Errorf("upload has not been created yet")
	}

	fileURL := fmt.Sprintf("%s/?id=%s", upload.client.URL, upload.ID())

	// Parse to get a nice escaped url
	return url.Parse(fileURL)
}

// DownloadZipArchive downloads all the upload files in a zip archive
func (upload *Upload) DownloadZipArchive() (reader io.ReadCloser, err error) {
	return upload.client.downloadArchive(upload.getParams())
}

// Delete remove the upload and all the associated files from the remote server
func (upload *Upload) Delete() (err error) {
	return upload.client.removeUpload(upload.getParams())
}
