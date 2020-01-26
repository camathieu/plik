package handlers

import (
	"fmt"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/plik/server/data"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GetFile download a file
func GetFile(ctx *juliet.Context, resp http.ResponseWriter, req *http.Request) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	// If a download domain is specified verify that the request comes from this specific domain
	if config.GetDownloadDomain() != nil {
		if req.Host != config.GetDownloadDomain().Host {
			downloadURL := fmt.Sprintf("%s://%s%s",
				config.GetDownloadDomain().Scheme,
				config.GetDownloadDomain().Host,
				req.RequestURI)
			log.Warningf("Invalid download domain %s, expected %s", req.Host, config.GetDownloadDomain().Host)
			http.Redirect(resp, req, downloadURL, http.StatusMovedPermanently)
			return
		}
	}

	// Get upload from context
	upload := context.GetUpload(ctx)
	if upload == nil {
		// This should never append
		log.Critical("Missing upload in getFileHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// Get file from context
	file := context.GetFile(ctx)
	if file == nil {
		// This should never append
		log.Critical("Missing file in getFileHandler")
		context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
		return
	}

	// File status pre-check
	if upload.Stream {
		if file.Status != common.FILE_UPLOADING {
			context.Fail(ctx, req, resp, fmt.Sprintf("file %s (%s) status is not %s", file.Name, file.ID, common.FILE_UPLOADING), http.StatusNotFound)
			return
		}
	} else {
		if file.Status != common.FILE_UPLOADED {
			context.Fail(ctx, req, resp, fmt.Sprintf("file %s (%s) status is not %s", file.Name, file.ID, common.FILE_UPLOADED), http.StatusNotFound)
			return
		}
	}

	if req.Method == "GET" && (upload.Stream || upload.OneShot) {
		// If this is a one shot or stream upload we have to ensure it's downloaded only once.
		tx := func(u *common.Upload) error {
			if u == nil {
				return fmt.Errorf("missing upload from transaction")
			}

			f, ok := u.Files[file.ID]
			if !ok {
				return fmt.Errorf("unable to find file %s (%s)", file.Name, file.ID)
			}

			// File status double-check
			if upload.Stream {
				if f.Status != common.FILE_UPLOADING {
					return common.NewTxError(fmt.Sprintf("invalid file %s (%s) status %s, expected %s", file.Name, file.ID, file.Status, common.FILE_UPLOADING), http.StatusBadRequest)
				}
				f.Status = common.FILE_DELETED
			} else if upload.OneShot {
				if f.Status != common.FILE_UPLOADED {
					return common.NewTxError(fmt.Sprintf("invalid file %s (%s) status %s, expected %s", file.Name, file.ID, file.Status, common.FILE_UPLOADED), http.StatusBadRequest)
				}
				f.Status = common.FILE_REMOVED
			}

			return nil
		}

		upload, err := context.GetMetadataBackend(ctx).UpdateUpload(upload, tx)
		if err != nil {
			if txError, ok := err.(common.TxError); ok {
				context.Fail(ctx, req, resp, txError.Error(), txError.GetStatusCode())
				return
			} else {
				log.Warningf("Unable to update upload metadata : %s", err)
				context.Fail(ctx, req, resp, "Unable to update upload metadata", http.StatusInternalServerError)
				return
			}
		}

		if !upload.Stream {
			// From now on we'll try to delete the file from the data backend whatever happens
			defer func() {
				err = DeleteRemovedFile(ctx, upload, file)
				if err != nil {
					log.Warningf("Unable to delete file %s (%s) : %s", file.Name, file.ID, err)
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
			backend = context.GetStreamBackend(ctx)
		} else {
			backend = context.GetDataBackend(ctx)
		}

		fileReader, err := backend.GetFile(upload, file.ID)
		if err != nil {
			log.Warningf("Failed to get file %s in upload %s : %s", file.Name, upload.ID, err)
			context.Fail(ctx, req, resp, fmt.Sprintf("Failed to read file %s", file.Name), http.StatusNotFound)
			return
		}
		defer func() { _ = fileReader.Close() }()

		// File is piped directly to http response body without buffering
		_, err = io.Copy(resp, fileReader)
		if err != nil {
			log.Warningf("Error while copying file to response : %s", err)
		}
	}
}
