/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package file

import (
	"fmt"
	"github.com/root-gg/plik/server/context"
	"io"
	"os"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
)

// Backend object
type Backend struct {
	Config *BackendConfig
}

// NewFileBackend instantiate a new File Data Backend
// from configuration passed as argument
func NewFileBackend(config *BackendConfig) (fb *Backend) {
	fb = new(Backend)
	fb.Config = config
	return
}

// GetFile implementation for file data backend will search
// on filesystem the asked file and return its reading filehandle
func (fb *Backend) GetFile(ctx *juliet.Context, upload *common.Upload, id string) (file io.ReadCloser, err error) {
	log := context.GetLogger(ctx)

	// Get upload directory
	directory, err := fb.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		log.Warningf("Unable to get upload directory : %s", err)
		return
	}

	// Get file path
	fullPath := directory + "/" + id

	// The file content will be piped directly
	// to the client response body
	file, err = os.Open(fullPath)
	if err != nil {
		err = log.EWarningf("Unable to open file %s : %s", fullPath, err)
		return
	}

	return
}

// AddFile implementation for file data backend will creates a new file for the given upload
// and save it on filesystem with the given file reader
func (fb *Backend) AddFile(ctx *juliet.Context, upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error) {
	log := context.GetLogger(ctx)

	// Get upload directory
	directory, err := fb.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		log.Warningf("Unable to get upload directory : %s", err)
		return
	}

	// Get file path
	fullPath := directory + "/" + file.ID

	// Create directory
	_, err = os.Stat(directory)
	if err != nil {
		err = os.MkdirAll(directory, 0777)
		if err != nil {
			err = log.EWarningf("Unable to create upload directory %s : %s", directory, err)
			return
		}
		log.Infof("Folder %s successfully created", directory)
	}

	// Create file
	out, err := os.Create(fullPath)
	if err != nil {
		err = log.EWarningf("Unable to create file %s : %s", fullPath, err)
		return
	}

	// Copy file data from the client request body
	// to the file system
	_, err = io.Copy(out, fileReader)
	if err != nil {
		err = log.EWarningf("Unable to save file %s : %s", fullPath, err)
		return
	}
	log.Infof("File %s successfully saved", fullPath)

	return
}

// RemoveFile implementation for file data backend will delete the given
// file from filesystem
func (fb *Backend) RemoveFile(ctx *juliet.Context, upload *common.Upload, id string) (err error) {
	log := context.GetLogger(ctx)

	// Get upload directory
	directory, err := fb.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		log.Warningf("Unable to get upload directory : %s", err)
		return
	}

	// Get file path
	fullPath := directory + "/" + id

	// Remove file
	err = os.Remove(fullPath)
	if err != nil {
		err = log.EWarningf("Unable to remove %s : %s", fullPath, err)
		return
	}

	log.Infof("File %s successfully removed", fullPath)

	return
}

// RemoveUpload implementation for file data backend will
// delete the whole upload. Given that an upload is a directory,
// we remove the whole directory at once.
func (fb *Backend) RemoveUpload(ctx *juliet.Context, upload *common.Upload) (err error) {
	log := context.GetLogger(ctx)

	// Get upload directory
	fullPath, err := fb.getDirectoryFromUploadID(upload.ID)
	if err != nil {
		log.Warningf("Unable to get upload directory : %s", err)
		return
	}

	// Remove everything at once
	err = os.RemoveAll(fullPath)
	if err != nil {
		err = log.EWarningf("Unable to remove %s : %s", fullPath, err)
		return
	}

	log.Infof("Upload %s successfully removed", fullPath)

	return
}

func (fb *Backend) getDirectoryFromUploadID(uploadID string) (string, error) {
	// To avoid too many files in the same directory
	// data directory is splitted in two levels the
	// first level is the 2 first chars from the upload id
	// it gives 3844 possibilities reaching 65535 files per
	// directory at ~250.000.000 files uploaded.

	if len(uploadID) < 3 {
		return "", fmt.Errorf("Invalid uploadid %s", uploadID)
	}
	return fb.Config.Directory + "/" + uploadID[:2] + "/" + uploadID, nil
}
