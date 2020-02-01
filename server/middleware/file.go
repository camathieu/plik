package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/root-gg/plik/server/context"
)

// File retrieve the requested file metadata from the metadataBackend and save it in the request context.
func File(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// Get upload from context
		upload := ctx.GetUpload()

		// Get the file id from the url params
		vars := mux.Vars(req)
		fileID := vars["fileID"]
		if fileID == "" {
			ctx.MissingParameter("file id")
			return
		}

		// Get the file name from the url params
		fileName := vars["filename"]
		if fileName == "" {
			ctx.MissingParameter("file name")
			return
		}

		// Get file object in upload metadata
		file, ok := upload.Files[fileID]
		if !ok {
			ctx.NotFound(fmt.Sprintf("file %s not found", fileID))
			return
		}

		// Compare url filename with upload filename
		if file.Name != fileName {
			ctx.InvalidParameter("file name")
			return
		}

		// Save file in the request context
		ctx.SetFile(file)

		next.ServeHTTP(resp, req)
	})
}
