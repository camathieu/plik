package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// GetArchive download all file of the upload in a zip archive
func GetArchive(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
	log := ctx.GetLogger()

	if !checkDownloadDomain(ctx) {
		return
	}

	// Get upload from context
	upload := ctx.GetUpload()

	if upload.Stream {
		ctx.Forbidden("archive feature is not available in stream mode")
		return
	}

	// Set content type
	resp.Header().Set("Content-Type", "application/zip")

	/* Additional security headers for possibly unsafe content */
	resp.Header().Set("X-Content-Type-Options", "nosniff")
	resp.Header().Set("X-XSS-Protection", "1; mode=block")
	resp.Header().Set("X-Frame-Options", "DENY")
	resp.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'; style-src 'none'; img-src 'none'; connect-src 'none'; font-src 'none'; object-src 'none'; media-src 'none'; child-src 'none'; form-action 'none'; frame-ancestors 'none'; plugin-types ''; sandbox ''")

	/* Additional header for disabling cache if the upload is OneShot */
	if upload.OneShot {
		resp.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
		resp.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0
		resp.Header().Set("Expires", "0")                                         // Proxies
	}

	// Get the file name from the url params
	vars := mux.Vars(req)
	fileName := vars["filename"]
	if fileName == "" {
		ctx.MissingParameter("file name")
	}

	if !strings.HasSuffix(fileName, ".zip") {
		ctx.InvalidParameter("file name, missing .zip extension")
	}

	// If "dl" GET params is set
	// -> Set Content-Disposition header
	// -> The client should download file instead of displaying it
	dl := req.URL.Query().Get("dl")
	if dl != "" {
		resp.Header().Set("Content-Disposition", fmt.Sprintf(`attachement; filename="%s"`, fileName))
	} else {
		resp.Header().Set("Content-Disposition", fmt.Sprintf(`filename="%s"`, fileName))
	}

	// HEAD Request => Do not print file, user just wants http headers
	// GET  Request => Print file content
	if req.Method == "GET" {
		// Get files to archive
		var files []*common.File

		if upload.OneShot {
			// If this is a one shot upload we have to ensure it's downloaded only once now
			tx := func(u *common.Upload) error {
				if u == nil {
					return common.NewHTTPError("upload does not exist anymore", http.StatusNotFound)
				}

				for _, f := range u.Files {
					// Ignore uploading, missing, removed, one shot already downloaded,...
					if f.Status != common.FileUploaded {
						continue
					}

					f.Status = common.FileRemoved
					files = append(files, f)
				}
				return nil
			}

			upload, err := ctx.GetMetadataBackend().UpdateUpload(upload, tx)
			if err != nil {
				handleTxError(ctx, "unable to update upload metadata", err)
				return
			}

			// From now on we'll try to delete the files from the data backend whatever happens
			defer func() {
				for _, file := range files {
					err = DeleteRemovedFile(ctx, upload, file)
					if err != nil {
						log.Warningf("Unable to delete file %s (%s) : %s", file.Name, file.ID, err)
					}
				}
			}()
		} else {
			// Without one shot mode we do not need as strong guaranties, no need to re-fetch upload metadata
			for _, f := range upload.Files {
				// Ignore uploading, missing, removed, one shot already downloaded,...
				if f.Status != common.FileUploaded {
					continue
				}

				files = append(files, f)
			}
		}

		if len(files) == 0 {
			ctx.BadRequest("nothing to archive")
			return
		}

		backend := ctx.GetDataBackend()

		// The zip archive is piped directly to http response body without buffering
		archive := zip.NewWriter(resp)

		for _, file := range files {
			fileReader, err := backend.GetFile(upload, file.ID)
			if err != nil {
				ctx.InternalServerError(fmt.Errorf("error retreiving file from data backend : %s", err))
				return
			}

			fileWriter, err := archive.Create(file.Name)
			if err != nil {
				ctx.InternalServerError(fmt.Errorf("error while creating zip archive : %s", err))
				return
			}

			// File is piped directly to zip archive thus to the http response body without buffering
			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				log.Warningf("error while copying zip archive to response body : %s", err)
			}

			err = fileReader.Close()
			if err != nil {
				log.Warningf("error while closing zip archive reader : %s", err)
			}
		}

		err := archive.Close()
		if err != nil {
			log.Warningf("error while closing zip archive : %s", err)
			return
		}
	}
}
