package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
)

// File retrieve the requested file metadata from the metadataBackend and save it in the request context.
func File(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := context.GetLogger(ctx)

		// Get upload from context
		upload := context.GetUpload(ctx)
		if upload == nil {
			// This should never append
			log.Critical("Missing upload in file middleware")
			context.Fail(ctx, req, resp, "Internal error", http.StatusInternalServerError)
			return
		}

		// Get the file id from the url params
		vars := mux.Vars(req)
		fileID := vars["fileID"]
		if fileID == "" {
			log.Warning("Missing file id")
			context.Fail(ctx, req, resp, "Missing file id", http.StatusBadRequest)
			return
		}

		// Get the file name from the url params
		fileName := vars["filename"]
		if fileName == "" {
			log.Warning("Missing file name")
			context.Fail(ctx, req, resp, "Missing file name", http.StatusBadRequest)
			return
		}

		// Get file object in upload metadata
		file, ok := upload.Files[fileID]
		if !ok {
			log.Warningf("File %s not found", fileID)
			context.Fail(ctx, req, resp, fmt.Sprintf("File %s not found", fileID), http.StatusNotFound)
			return
		}

		// Compare url filename with upload filename
		if file.Name != fileName {
			log.Warningf("Invalid filename %s mismatch %s", fileName, file.Name)
			context.Fail(ctx, req, resp, fmt.Sprintf("File %s not found", fileName), http.StatusNotFound)
			return
		}

		// Save file in the request context
		context.SetFile(ctx, file)

		next.ServeHTTP(resp, req)
	})
}
