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

package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/utils"
)

type preprocessOutputReturn struct {
	size     int64
	md5sum   string
	mimeType string
	err      error
}

// AddFile add a file to an existing upload.
func AddFile(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := common.GetLogger(ctx)

	// Get upload from context
	upload := common.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in AddFileHandler")
		common.Fail(ctx, req, resp, "Internal error", 500)
		return
	}

	// Check anonymous user uploads
	if common.Config.NoAnonymousUploads {
		user := common.GetUser(ctx)
		if user == nil {
			log.Warning("Unable to add file from anonymous user")
			common.Fail(ctx, req, resp, "Unable to add file from anonymous user. Please login or use a cli token.", 403)
			return
		}
	}

	// Check authorization
	if !upload.IsAdmin {
		log.Warningf("Unable to add file : unauthorized")
		common.Fail(ctx, req, resp, "You are not allowed to add file to this upload", 403)
		return
	}

	// Get the file id from the url params
	vars := mux.Vars(req)
	fileID := vars["fileID"]

	var newFile *common.File
	if fileID == "" {
		// Limit number of files per upload
		if len(upload.Files) >= common.Config.MaxFilePerUpload {
			err := log.EWarningf("Unable to add file : Maximum number file per upload reached (%d)", common.Config.MaxFilePerUpload)
			common.Fail(ctx, req, resp, err.Error(), 403)
			return
		}

		// Create a new file object
		newFile = common.NewFile()
		newFile.Type = "application/octet-stream"
	} else {
		// Get file object from upload
		if _, ok := upload.Files[fileID]; ok {
			newFile = upload.Files[fileID]
		} else {
			log.Warningf("Invalid file id %s", fileID)
			common.Fail(ctx, req, resp, "Invalid file id", 404)
			return
		}
	}

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, newFile.ID)
	log.SetPrefix(prefix)

	// Save file to the context
	ctx.Set("file", newFile)

	// Get file handle from multipart request
	var file io.Reader
	multiPartReader, err := req.MultipartReader()
	if err != nil {
		log.Warningf("Failed to get file from multipart request : %s", err)
		common.Fail(ctx, req, resp, "Failed to get file from multipart request", 400)
		return
	}

	// Read multipart body until the "file" part
	for {
		part, errPart := multiPartReader.NextPart()
		if errPart == io.EOF {
			break
		}
		if part.FormName() == "file" {
			file = part

			// Check file name length
			if len(part.FileName()) > 1024 {
				log.Warning("File name is too long")
				common.Fail(ctx, req, resp, "File name is too long. Maximum length is 1024 characters", 400)
				return
			}

			newFile.Name = part.FileName()
			break
		}
	}
	if file == nil {
		log.Warning("Missing file from multipart request")
		common.Fail(ctx, req, resp, "Missing file from multipart request", 400)
		return
	}
	if newFile.Name == "" {
		log.Warning("Missing file name from multipart request")
		common.Fail(ctx, req, resp, "Missing file name from multipart request", 400)
		return
	}

	// Update request logger prefix
	prefix = fmt.Sprintf("%s[%s]", log.Prefix, newFile.Name)
	log.SetPrefix(prefix)

	// Pipe file data from the request body to a preprocessing goroutine
	//  - Guess content type
	//  - Compute/Limit upload size
	//  - Compute md5sum
	preprocessReader, preprocessWriter := io.Pipe()
	preprocessOutputCh := make(chan preprocessOutputReturn)
	go preprocessor(ctx, file, preprocessWriter, preprocessOutputCh)

	// Save file in the data backend
	var backend data.Backend
	if upload.Stream {
		backend = data.GetStreamBackend()
	} else {
		backend = data.GetDataBackend()
	}
	backendDetails, err := backend.AddFile(ctx, upload, newFile, preprocessReader)
	if err != nil {
		log.Warningf("Unable to save file : %s", err)
		common.Fail(ctx, req, resp, "Unable to save file", 500)
		return
	}

	// Get preprocessor goroutine output
	preprocessOutput := <-preprocessOutputCh
	if preprocessOutput.err != nil {
		log.Warningf("Unable to execute preprocessor : %s", err)
		common.Fail(ctx, req, resp, "Unable to save file", 500)
		return
	}

	// Fill-in file information
	newFile.Type = preprocessOutput.mimeType
	newFile.CurrentSize = preprocessOutput.size
	newFile.Md5 = preprocessOutput.md5sum

	if upload.Stream {
		newFile.Status = "downloaded"
	} else {
		newFile.Status = "uploaded"
	}
	newFile.UploadDate = time.Now().Unix()
	newFile.BackendDetails = backendDetails

	// Update upload metadata
	upload.Files[newFile.ID] = newFile
	err = metadata.GetMetaDataBackend().AddOrUpdateFile(ctx, upload, newFile)
	if err != nil {
		log.Warningf("Unable to update metadata : %s", err)
		common.Fail(ctx, req, resp, "Unable to update upload metadata", 500)
		return
	}

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	newFile.Sanitize()

	if common.IsQuick(ctx) {
		// Print the file url in the response.
		var url string
		if common.Config.DownloadDomainURL != nil {
			url = fmt.Sprintf("%s://%s", common.Config.DownloadDomainURL.Scheme, common.Config.DownloadDomainURL.Host)
		} else {
			// This will most likely break behind a reverse proxy
			proto := "http"
			if req.TLS != nil {
				proto = "https"
			}
			url = fmt.Sprintf("%s://%s", proto, req.Host)
		}

		url += fmt.Sprintf("/file/%s/%s/%s", upload.ID, newFile.ID, newFile.Name)

		resp.Write([]byte(url + "\n"))
	} else {
		// Print file metadata in json in the response.
		var json []byte
		if json, err = utils.ToJson(newFile); err == nil {
			resp.Write(json)
		} else {
			log.Warningf("Unable to serialize json response : %s", err)
			common.Fail(ctx, req, resp, "Unable to serialize json response", 500)
		}
	}
}

//  - Guess content type
//  - Compute/Limit upload size
//  - Compute md5sum
func preprocessor(ctx *juliet.Context, file io.Reader, preprocessWriter io.WriteCloser, outputCh chan preprocessOutputReturn) {
	log := common.GetLogger(ctx)

	var err error
	var totalBytes int64
	var mimeType string
	var md5sum string

	md5Hash := md5.New()
	buf := make([]byte, 1048)

	eof := false
	for !eof {
		bytesRead, err := file.Read(buf)
		if err == io.EOF {
			eof = true
			if bytesRead <= 0 {
				break
			}
		} else if err != nil {
			err = log.EWarningf("Unable to read data from request body : %s", err)
			break
		}

		// Detect the content-type using the 512 first bytes
		if totalBytes == 0 {
			mimeType = http.DetectContentType(buf)
		}

		// Increment size
		totalBytes += int64(bytesRead)

		// Check upload max size limit
		if int64(totalBytes) > common.Config.MaxFileSize {
			err = log.EWarningf("File too big (limit is set to %d bytes)", common.Config.MaxFileSize)
			break
		}

		// Compute md5sum
		md5Hash.Write(buf[:bytesRead])

		// Forward data to the data backend
		bytesWritten, err := preprocessWriter.Write(buf[:bytesRead])
		if err != nil {
			log.Warning(err.Error())
			break
		}
		if bytesWritten != bytesRead {
			err = log.EWarningf("Invalid number of bytes written. Expected %d but got %d", bytesRead, bytesWritten)
			break
		}
	}

	errClose := preprocessWriter.Close()
	if errClose != nil {
		log.Warningf("Unable to close preprocessWriter : %s", err)
	}

	if err != nil {
		outputCh <- preprocessOutputReturn{err: err}
	} else {
		md5sum = fmt.Sprintf("%x", md5Hash.Sum(nil))
		outputCh <- preprocessOutputReturn{size: totalBytes, md5sum: md5sum, mimeType: mimeType}
	}

	close(outputCh)
}
