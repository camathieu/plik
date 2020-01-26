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
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/plik/server/data"
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
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	// Get upload from context
	upload := context.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in AddFileHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check anonymous user uploads
	if config.NoAnonymousUploads {
		user := context.GetUser(ctx)
		if user == nil {
			log.Warning("Unable to add file from anonymous user")
			context.Fail(ctx, req, resp, "Unable to add file from anonymous user. Please login or use a cli token.", http.StatusForbidden)
			return
		}
	}

	// Check authorization
	if !context.IsUploadAdmin(ctx) {
		log.Warningf("Unable to add file : unauthorized")
		context.Fail(ctx, req, resp, "You are not allowed to add file to this upload", http.StatusForbidden)
		return
	}

	// Get the file id from the url params
	vars := mux.Vars(req)
	fileID := vars["fileID"]

	var file *common.File
	if fileID == "" {
		// TODO we keep this for backward compatibility
		// TODO the best way would be to have a separate handler to create file metadata and then to call this one
		// TODO especially now that we have a way to quick upload in one go for simple use cases

		// Create a new file object
		file = common.NewFile()
		file.Type = "application/octet-stream"
		file.Status = common.FileMissing
	} else {
		// Get file object from upload
		var ok bool
		file, ok = upload.Files[fileID]
		if !ok {
			log.Warningf("Missing file with id %s", fileID)
			context.Fail(ctx, req, resp, "Invalid file id", http.StatusNotFound)
			return
		}
	}

	// Check file status and set to FileUploading
	// This avoids file overwriting
	tx := func(u *common.Upload) (err error) {
		if u == nil {
			return fmt.Errorf("missing upload from transaction")
		}

		f, ok := u.Files[file.ID]
		if !ok {
			// Add new file to the upload metadata
			u.Files[file.ID] = file
			f = file
		}

		// Limit number of files per upload
		if len(u.Files) > config.MaxFilePerUpload {
			return common.NewTxError(fmt.Sprintf("Maximum number file per upload reached (%d)", config.MaxFilePerUpload), http.StatusBadRequest)
		}

		if !(f.Status == "" || f.Status == common.FileMissing) {
			return common.NewTxError(fmt.Sprintf("File %s (%s) has already been uploaded or removed", f.Name, f.ID), http.StatusBadRequest)
		}

		f.Status = common.FileUploading

		return nil
	}

	// Update upload metadata
	upload, err := context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
	if err != nil {
		if txError, ok := err.(common.TxError); ok {
			context.Fail(ctx, req, resp, txError.Error(), txError.GetStatusCode())
		} else {
			log.Warningf("Unable to update upload : %s", err)
			context.Fail(ctx, req, resp, "Unable to add file", http.StatusInternalServerError)
		}
		return
	}

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, file.ID)
	log.SetPrefix(prefix)

	// Get file handle from multipart request
	var fileReader io.Reader
	multiPartReader, err := req.MultipartReader()
	if err != nil {
		log.Warningf("Failed to get file from multipart request : %s", err)
		context.Fail(ctx, req, resp, "Failed to get file from multipart request", http.StatusBadRequest)
		return
	}

	// Read multipart body until the "file" part
	for {
		part, errPart := multiPartReader.NextPart()
		if errPart == io.EOF {
			break
		}
		if errPart != nil {
			log.Warningf("multipart reader error : %v", errPart)
			context.Fail(ctx, req, resp, "Unable to read file", http.StatusBadRequest)
			return
		}
		if part.FormName() == "file" {
			fileReader = part

			if file.Name != "" {
				if file.Name != part.FileName() {
					context.Fail(ctx, req, resp, fmt.Sprintf("Invalid filename %s, expected %s", part.FileName(), file.Name), http.StatusBadRequest)
					return
				}
			} else {
				// Check file name length
				if len(part.FileName()) > 1024 {
					context.Fail(ctx, req, resp, "File name is too long. Maximum length is 1024 characters", http.StatusBadRequest)
					return
				}

				file.Name = part.FileName()
			}

			break
		}
	}
	if fileReader == nil {
		context.Fail(ctx, req, resp, "Missing file from multipart request", http.StatusBadRequest)
		return
	}
	if file.Name == "" {
		context.Fail(ctx, req, resp, "Missing file name from multipart request", http.StatusBadRequest)
		return
	}

	// Update request logger prefix
	prefix = fmt.Sprintf("%s[%s]", log.Prefix, file.Name)
	log.SetPrefix(prefix)

	// Pipe file data from the request body to a preprocessing goroutine
	//  - Guess content type
	//  - Compute/Limit upload size
	//  - Compute md5sum
	preprocessReader, preprocessWriter := io.Pipe()
	preprocessOutputCh := make(chan preprocessOutputReturn)
	go preprocessor(ctx, fileReader, preprocessWriter, preprocessOutputCh)

	// Save file in the data backend
	var backend data.Backend
	if upload.Stream {
		backend = context.GetStreamBackend(ctx)
	} else {
		backend = context.GetDataBackend(ctx)
	}

	backendDetails, err := backend.AddFile(upload, file, preprocessReader)
	if err != nil {
		log.Warningf("Unable to save file : %s", err)
		context.Fail(ctx, req, resp, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Get preprocessor goroutine output
	preprocessOutput := <-preprocessOutputCh
	if preprocessOutput.err != nil {
		log.Warningf("Unable to execute preprocessor : %s", preprocessOutput.err)
		context.Fail(ctx, req, resp, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Fill-in file information
	file.Type = preprocessOutput.mimeType
	file.CurrentSize = preprocessOutput.size
	file.Md5 = preprocessOutput.md5sum
	file.UploadDate = time.Now().Unix()
	file.BackendDetails = backendDetails

	if !upload.Stream {
		// Double check file status and update upload metadata
		tx = func(u *common.Upload) (err error) {
			if u == nil {
				return fmt.Errorf("missing upload from upload transaction")
			}
			// Just to check that the file has not been already removed
			f, ok := u.Files[file.ID]
			if !ok {
				return fmt.Errorf("missing file %s from upload transaction", file.ID)
			}
			if f.Status != common.FileUploading {
				return common.NewTxError(fmt.Sprintf("invalid file status %s, expected %s", f.Status, common.FileUploading), http.StatusInternalServerError)
			}

			// Update file status
			file.Status = common.FileUploaded

			// Update file
			u.Files[file.ID] = file

			return nil
		}

		upload, err = context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
		if err != nil {
			if txError, ok := err.(common.TxError); ok {
				context.Fail(ctx, req, resp, txError.Error(), txError.GetStatusCode())
			} else {
				log.Warningf("Unable to update upload metadata : %s", err)
				context.Fail(ctx, req, resp, "Unable to update upload metadata", http.StatusInternalServerError)
			}
			return
		}

		var ok bool
		file, ok = upload.Files[file.ID]
		if !ok {
			log.Warningf("Missing file from upload after metadata update, wtf ?!")
			context.Fail(ctx, req, resp, "Missing file from upload after metadata update", http.StatusInternalServerError)
			return
		}
	} else {
		// For steam upload the file will be removed by the getFile handler
		// If there is only one file the upload will be removed too
		file.Status = common.FileDeleted
	}

	// Remove all private information (ip, data backend details, ...) before
	// sending metadata back to the client
	file.Sanitize()

	if context.IsQuick(ctx) {
		// Print the file url in the response.
		var url string
		if context.GetConfig(ctx).GetDownloadDomain() != nil {
			url = context.GetConfig(ctx).GetDownloadDomain().String()
		} else {
			// This will break behind any non transparent reverse proxy
			proto := "http"
			if req.TLS != nil {
				proto = "https"
			}
			url = fmt.Sprintf("%s://%s", proto, req.Host)
		}

		url += fmt.Sprintf("/file/%s/%s/%s", upload.ID, file.ID, file.Name)

		_, _ = resp.Write([]byte(url + "\n"))
	} else {
		// Print file metadata in json in the response.
		var json []byte
		if json, err = utils.ToJson(file); err == nil {
			_, _ = resp.Write(json)
		} else {
			log.Warningf("Unable to serialize json response : %s", err)
			context.Fail(ctx, req, resp, "Unable to serialize json response", http.StatusInternalServerError)
			return
		}
	}
}

//  - Guess content type
//  - Compute/Limit upload size
//  - Compute md5sum
func preprocessor(ctx *juliet.Context, file io.Reader, preprocessWriter io.WriteCloser, outputCh chan preprocessOutputReturn) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	var err error
	var totalBytes int64
	var mimeType string
	var md5sum string

	md5Hash := md5.New()
	buf := make([]byte, 1048)

	eof := false
	for !eof {
		bytesRead := 0
		bytesRead, err = file.Read(buf)
		if err == io.EOF {
			eof = true
			err = nil
			if bytesRead <= 0 {
				break
			}
		} else if err != nil {
			err = fmt.Errorf("unable to read data from request body : %s", err)
			break
		}

		// Detect the content-type using the 512 first bytes
		if totalBytes == 0 {
			mimeType = http.DetectContentType(buf)
		}

		// Increment size
		totalBytes += int64(bytesRead)

		// Check upload max size limit
		if int64(totalBytes) > config.MaxFileSize {
			err = fmt.Errorf("file too big (limit is set to %d bytes)", config.MaxFileSize)
			break
		}

		// Compute md5sum
		_, err = md5Hash.Write(buf[:bytesRead])
		if err != nil {
			err = fmt.Errorf(err.Error())
			break
		}

		// Forward data to the data backend
		bytesWritten, err := preprocessWriter.Write(buf[:bytesRead])
		if err != nil {
			err = fmt.Errorf(err.Error())
			break
		}
		if bytesWritten != bytesRead {
			err = fmt.Errorf("invalid number of bytes written. Expected %d but got %d", bytesRead, bytesWritten)
			break
		}
	}

	errClose := preprocessWriter.Close()
	if errClose != nil {
		log.Warningf("unable to close preprocessWriter : %s", err)
	}

	if err != nil {
		outputCh <- preprocessOutputReturn{err: err}
	} else {
		md5sum = fmt.Sprintf("%x", md5Hash.Sum(nil))
		outputCh <- preprocessOutputReturn{size: totalBytes, md5sum: md5sum, mimeType: mimeType}
	}

	close(outputCh)
}
