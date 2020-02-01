package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

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
func AddFile(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()
	config := ctx.GetConfig()

	// Get upload from context
	upload := ctx.GetUpload()
	if upload == nil {
		ctx.InternalServerError("missing upload from context", nil)
		return
	}

	// Check authorization
	if !ctx.IsUploadAdmin() {
		ctx.Forbidden("you are not allowed to add file to this upload")
		return
	}

	// Get the file id from the url params
	vars := mux.Vars(req)
	fileID := vars["fileID"]

	var file *common.File
	if fileID == "" {
		// Create a new file object
		file = common.NewFile()
		file.Type = "application/octet-stream"
		file.Status = common.FileMissing
	} else {
		// Get file object from upload
		var ok bool
		file, ok = upload.Files[fileID]
		if !ok {
			ctx.NotFound("file %s not found", fileID)
			return
		}
	}

	// Check file status and set to FileUploading
	// This avoids file overwriting
	tx := func(u *common.Upload) (err error) {
		if u == nil {
			return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
		}

		f, ok := u.Files[file.ID]
		if !ok {
			// Add new file to the upload metadata
			u.Files[file.ID] = file
			f = file
		}

		// Limit number of files per upload
		if len(u.Files) > config.MaxFilePerUpload {
			return common.NewHTTPError(fmt.Sprintf("maximum number file per upload reached, limit is %d", config.MaxFilePerUpload), http.StatusBadRequest)
		}

		if !(f.Status == "" || f.Status == common.FileMissing) {
			return common.NewHTTPError(fmt.Sprintf("file %s (%s) has already been uploaded or removed", f.Name, f.ID), http.StatusBadRequest)
		}

		f.Status = common.FileUploading

		return nil
	}

	// Update upload metadata
	upload, err := ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		handleTxError(ctx, "unable to update upload metadata", err)
		return
	}

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, file.ID)
	log.SetPrefix(prefix)

	// Get file handle from multipart request
	var fileReader io.Reader
	multiPartReader, err := req.MultipartReader()
	if err != nil {
		ctx.InvalidParameter("multipart form : %s", err)
		return
	}

	// Read multipart body until the "file" part
	for {
		part, errPart := multiPartReader.NextPart()
		if errPart == io.EOF {
			break
		}
		if errPart != nil {
			ctx.InvalidParameter("multipart form : %s", errPart)
			return
		}
		if part.FormName() == "file" {
			fileReader = part

			if file.Name != "" {
				if file.Name != part.FileName() {
					ctx.BadRequest("invalid filename %s, expected %s", part.FileName(), file.Name)
					return
				}
			} else {
				// Check file name length
				if len(part.FileName()) > 1024 {
					ctx.BadRequest("file name is too long, maximum allowed length is 1024 characters")
					return
				}

				file.Name = part.FileName()
			}

			break
		}
	}
	if fileReader == nil {
		ctx.MissingParameter("file")
		return
	}
	if file.Name == "" {
		ctx.MissingParameter("file name")
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
		backend = ctx.GetStreamBackend()
	} else {
		backend = ctx.GetDataBackend()
	}

	backendDetails, err := backend.AddFile(upload, file, preprocessReader)
	if err != nil {
		ctx.InternalServerError("unable to save file", err)
		return
	}

	// Get preprocessor goroutine output
	preprocessOutput := <-preprocessOutputCh
	if preprocessOutput.err != nil {
		handleTxError(ctx, "unable to execute preprocessor", preprocessOutput.err)
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
				return common.NewHTTPError("missing upload from upload transaction", http.StatusNotFound)
			}

			// Just to check that the file has not been already removed
			f, ok := u.Files[file.ID]
			if !ok {
				return fmt.Errorf("missing file %s from upload transaction", file.ID)
			}
			if f.Status != common.FileUploading {
				return common.NewHTTPError(fmt.Sprintf("invalid file status %s, expected %s", f.Status, common.FileUploading), http.StatusInternalServerError)
			}

			// Update file status
			file.Status = common.FileUploaded

			// Update file
			u.Files[file.ID] = file

			return nil
		}

		upload, err = ctx.GetMetadataBackend().UpdateUpload(upload, tx)
		if err != nil {
			handleTxError(ctx, "unable to update upload metadata", err)
			return
		}

		// Get updated file
		var ok bool
		file, ok = upload.Files[file.ID]
		if !ok {
			ctx.InternalServerError("missing file from upload after metadata update", nil)
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

	if ctx.IsQuick() {
		// Print the file url in the response.
		var url string
		if ctx.GetConfig().GetDownloadDomain() != nil {
			url = ctx.GetConfig().GetDownloadDomain().String()
		} else {
			ctx.BadRequest("download domain must be set server side for quick upload to work")
			return
		}

		url += fmt.Sprintf("/file/%s/%s/%s", upload.ID, file.ID, file.Name)

		_, _ = resp.Write([]byte(url + "\n"))
	} else {
		// Print file metadata in json in the response.
		var json []byte
		if json, err = utils.ToJson(file); err == nil {
			_, _ = resp.Write(json)
		} else {
			ctx.InternalServerError("unable to serialize json response", err)
			return
		}
	}
}

//  - Guess content type
//  - Compute/Limit upload size
//  - Compute md5sum
func preprocessor(ctx *context.Context, file io.Reader, preprocessWriter io.WriteCloser, outputCh chan preprocessOutputReturn) {
	log := ctx.GetLogger()
	config := ctx.GetConfig()

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
			err = common.NewHTTPError(fmt.Sprintf("unable to read data from request body : %s", err), http.StatusInternalServerError)
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
			err = common.NewHTTPError(fmt.Sprintf("file too big (limit is set to %d bytes)", config.MaxFileSize), http.StatusBadRequest)
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
