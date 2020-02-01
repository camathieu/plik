package handlers

import (
	"fmt"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/plik/server/data"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GetFile download a file
func GetFile(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()

	if !checkDownloadDomain(ctx) {
		return
	}

	// Get upload from context
	upload := ctx.GetUpload()

	// Get file from context
	file := ctx.GetFile()

	// File status pre-check
	if upload.Stream {
		if file.Status != common.FileUploading {
			ctx.NotFound(fmt.Sprintf("file %s (%s) is not available : %s", file.Name, file.ID, file.Status))
			return
		}
	} else {
		if file.Status != common.FileUploaded {
			ctx.NotFound(fmt.Sprintf("file %s (%s) is not available : %s", file.Name, file.ID, file.Status))
			return
		}
	}

	if req.Method == "GET" && (upload.Stream || upload.OneShot) {
		// If this is a one shot or stream upload we have to ensure it's downloaded only once.
		tx := func(u *common.Upload) error {
			if u == nil {
				return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
			}

			f, ok := u.Files[file.ID]
			if !ok {
				return fmt.Errorf("unable to find file %s (%s)", file.Name, file.ID)
			}

			// File status double-check
			if upload.Stream {
				if f.Status != common.FileUploading {
					return common.NewHTTPError(fmt.Sprintf("invalid file %s (%s) status %s, expected %s", file.Name, file.ID, file.Status, common.FileUploading), http.StatusBadRequest)
				}
				f.Status = common.FileDeleted
			} else if upload.OneShot {
				if f.Status != common.FileUploaded {
					return common.NewHTTPError(fmt.Sprintf("invalid file %s (%s) status %s, expected %s", file.Name, file.ID, file.Status, common.FileUploaded), http.StatusBadRequest)
				}
				f.Status = common.FileRemoved
			}

			return nil
		}

		upload, err := ctx.GetMetadataBackend().UpdateUpload(upload, tx)
		if err != nil {
			handleTxError(ctx, "unable to update upload metadata", err)
			return
		}

		if !upload.Stream {
			// From now on we'll try to delete the file from the data backend whatever happens
			defer func() {
				err = DeleteRemovedFile(ctx, upload, file)
				if err != nil {
					log.Warningf("enable to delete file %s (%s) : %s", file.Name, file.ID, err)
				}
			}()
		}
	}

	// Avoid rendering HTML in browser
	if strings.Contains(file.Type, "html") {
		file.Type = "text/plain"
	}

	// Force the download of the following types as they are blocked by the CSP Header and won't display properly.
	if file.Type == "" || strings.Contains(file.Type, "flash") || strings.Contains(file.Type, "pdf") {
		file.Type = "application/octet-stream"
	}

	// Set content type and print file
	resp.Header().Set("Content-Type", file.Type)

	/* Additional security headers for possibly unsafe content */
	resp.Header().Set("X-Content-Type-Options", "nosniff")
	resp.Header().Set("X-XSS-Protection", "1; mode=block")
	resp.Header().Set("X-Frame-Options", "DENY")
	resp.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'; style-src 'none'; img-src 'none'; connect-src 'none'; font-src 'none'; object-src 'none'; media-src 'self'; child-src 'none'; form-action 'none'; frame-ancestors 'none'; plugin-types; sandbox")

	/* Additional header for disabling cache if the upload is OneShot */
	if upload.OneShot { // If this is a one shot or stream upload we have to ensure it's downloaded only once.
		resp.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
		resp.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0
		resp.Header().Set("Expires", "0")                                         // Proxies
	}

	if file.CurrentSize > 0 {
		resp.Header().Set("Content-Length", strconv.Itoa(int(file.CurrentSize)))
	}

	// If "dl" GET params is set
	// -> Set Content-Disposition header
	// -> The client should download file instead of displaying it
	dl := req.URL.Query().Get("dl")
	if dl != "" {
		resp.Header().Set("Content-Disposition", fmt.Sprintf(`attachement; filename="%s"`, file.Name))
	} else {
		resp.Header().Set("Content-Disposition", fmt.Sprintf(`filename="%s"`, file.Name))
	}

	// HEAD Request => Do not print file, user just wants http headers
	// GET  Request => Print file content
	if req.Method == "GET" {
		// Get file in data backend
		var backend data.Backend
		if upload.Stream {
			backend = ctx.GetStreamBackend()
		} else {
			backend = ctx.GetDataBackend()
		}

		fileReader, err := backend.GetFile(upload, file.ID)
		if err != nil {
			ctx.InternalServerError(fmt.Errorf("error retreiving file from data backend : %s", err))
			return
		}
		defer func() { _ = fileReader.Close() }()

		// File is piped directly to http response body without buffering
		_, err = io.Copy(resp, fileReader)
		if err != nil {
			log.Warningf("error while copying file to response : %s", err)
		}
	}
}
