package handlers

import (
	"crypto/md5"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"time"

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
		panic("missing upload from context")
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

	// Update request logger prefix
	prefix := fmt.Sprintf("%s[%s]", log.Prefix, file.ID)
	log.SetPrefix(prefix)

	// Get file handle form multipart request
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
		ctx.MissingParameter("file from multipart form")
		return
	}
	if file.Name == "" {
		ctx.MissingParameter("file name from multipart form")
		return
	}

	// Update request logger prefix
	prefix = fmt.Sprintf("%s[%s]", log.Prefix, file.Name)
	log.SetPrefix(prefix)

	// Check file status and set status to common.FileUploading to avoid multiple uploads in parallel
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
			return common.NewHTTPError(fmt.Sprintf("file status is %s", f.Status), http.StatusBadRequest)
		}

		f.Status = common.FileUploading

		return nil
	}

	// Update upload metadata
	upload, err = ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		handleTxError(ctx, "unable to update upload metadata", err)
		return
	}

	// Get updated file
	var ok bool
	file, ok = upload.Files[file.ID]
	if !ok {
		panic("missing file from upload after metadata update")
	}

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
		// TODO : file status is left to common.FileUploading we should set it to some common.FileUploadError
		// TODO : or we can set it back to common.FileMissing if we are sure data backends will handle that
		ctx.InternalServerError("unable to save file", err)
		return
	}

	// Get preprocessor goroutine output
	preprocessOutput := <-preprocessOutputCh
	if preprocessOutput.err != nil {
		// TODO : file status is left to common.FileUploading we should set it to some common.FileUploadError
		// TODO : or we can set it back to common.FileMissing if we are sure data backends will handle that
		handleTxError(ctx, "unable to execute preprocessor", preprocessOutput.err)
		return
	}

	// update upload metadata
	tx = func(u *common.Upload) (err error) {
		if u == nil {
			return common.NewHTTPError("upload does not exists anymore", http.StatusNotFound)
		}

		// This should never happen as the file status should not be changed until it has been set to uploaded
		f, ok := u.Files[file.ID]
		if !ok {
			return fmt.Errorf("missing file %s from upload transaction", file.ID)
		}

		if u.Stream {
			if f.Status != common.FileUploading {
				return fmt.Errorf("invalid file status %s, expected %s", f.Status, common.FileUploading)
			}
		}

		// Update file status
		if upload.Stream {
			f.Status = common.FileDeleted
		} else {
			f.Status = common.FileUploaded
		}

		// Fill-in file information
		f.Type = preprocessOutput.mimeType
		f.CurrentSize = preprocessOutput.size
		f.Md5 = preprocessOutput.md5sum
		f.UploadDate = time.Now().Unix()
		f.BackendDetails = backendDetails

		return nil
	}

	// Update upload metadata
	upload, err = ctx.GetMetadataBackend().UpdateUpload(upload, tx)
	if err != nil {
		handleTxError(ctx, "unable to update upload metadata", err)
		return
	}

	// Get updated file
	var ok2 bool
	file, ok2 = upload.Files[file.ID]
	if !ok2 {
		ctx.InternalServerError("missing file from upload after metadata update", nil)
		return
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
			url = ctx.GetConfig().GetServerURL().String()
		}

		url += fmt.Sprintf("/file/%s/%s/%s", upload.ID, file.ID, file.Name)

		_, _ = resp.Write([]byte(url + "\n"))
	} else {
		// Print file metadata in json in the response.
		var json []byte
		if json, err = utils.ToJson(file); err == nil {
			_, _ = resp.Write(json)
		} else {
			panic(fmt.Errorf("unable to serialize json response : %s", err))
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
