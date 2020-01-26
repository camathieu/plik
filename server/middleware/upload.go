package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/utils"
)

// Upload retrieve the requested upload metadata from the metadataBackend and save it to the request context.
func Upload(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := context.GetLogger(ctx)

		// Get the upload id from the url params
		vars := mux.Vars(req)
		uploadID := vars["uploadID"]
		if uploadID == "" {
			log.Warning("Missing upload id")
			context.Fail(ctx, req, resp, "Missing upload id", http.StatusBadRequest)
			return
		}

		// Get upload metadata
		upload, err := context.GetMetadataBackend(ctx).GetUpload(uploadID)
		if err != nil {
			log.Warningf("Unable to get upload metadata : %s", err)
			context.Fail(ctx, req, resp, fmt.Sprintf("Unable to get upload metadata : %s", err), http.StatusInternalServerError)
			return
		}
		if upload == nil {
			context.Fail(ctx, req, resp, fmt.Sprintf("Upload %s not found", uploadID), http.StatusNotFound)
			return
		}

		// Update request logger prefix
		prefix := fmt.Sprintf("%s[%s]", log.Prefix, uploadID)
		log.SetPrefix(prefix)

		// Test if upload is not expired
		if upload.IsExpired() {
			log.Warningf("Upload is expired since %s", time.Since(time.Unix(upload.Creation, int64(0)).Add(time.Duration(upload.TTL)*time.Second)).String())
			context.Fail(ctx, req, resp, fmt.Sprintf("Upload %s has expired", uploadID), http.StatusNotFound)
			return
		}

		// Save upload in the request context
		context.SetUpload(ctx, upload)

		forbidden := func() {
			resp.Header().Set("WWW-Authenticate", "Basic realm=\"plik\"")

			// Shouldn't redirect here to let the browser ask for credentials and retry
			context.SetRedirectOnFailure(ctx, false)

			context.Fail(ctx, req, resp, "Please provide valid credentials to access this upload", http.StatusUnauthorized)
		}

		// Check upload token
		uploadToken := req.Header.Get("X-UploadToken")
		if uploadToken != "" && uploadToken == upload.UploadToken {
			context.SetUploadAdmin(ctx, true)
		} else {
			token := context.GetToken(ctx)
			if token != nil {
				// A user authenticated with a token can manage uploads created with such token
				if upload.Token == token.Token {
					context.SetUploadAdmin(ctx, true)
				}
			} else {
				// Check if upload belongs to user or if user is admin
				user := context.GetUser(ctx)
				if context.IsAdmin(ctx) {
					context.SetUploadAdmin(ctx, true)
				} else if user != nil && upload.User == user.ID {
					context.SetUploadAdmin(ctx, true)
				}
			}
		}

		// Handle basic auth if upload is password protected
		if upload.ProtectedByPassword && !context.IsUploadAdmin(ctx) {
			if req.Header.Get("Authorization") == "" {
				log.Warning("Missing Authorization header")
				forbidden()
				return
			}

			// Basic auth Authorization header must be set to
			// "Basic base64("login:password")". Only the md5sum
			// of the base64 string is saved in the upload metadata
			auth := strings.Split(req.Header.Get("Authorization"), " ")
			if len(auth) != 2 {
				log.Warningf("Inavlid Authorization header %s", req.Header.Get("Authorization"))
				forbidden()
				return
			}
			if auth[0] != "Basic" {
				log.Warningf("Inavlid http authorization scheme : %s", auth[0])
				forbidden()
				return
			}
			var md5sum string
			md5sum, err = utils.Md5sum(auth[1])
			if err != nil {
				log.Warningf("Unable to hash credentials : %s", err)
				forbidden()
				return
			}
			if md5sum != upload.Password {
				log.Warning("Invalid credentials")
				forbidden()
				return
			}
		}

		next.ServeHTTP(resp, req)
	})
}
