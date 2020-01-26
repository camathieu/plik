package middleware

import (
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// CreateUpload create a new upload on the fly to be used in the next handler
func CreateUpload(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		upload := common.NewUpload()
		upload.Create()

		// Set upload remote IP
		upload.RemoteIP = context.GetSourceIP(ctx).String()

		// Set upload one shot mode
		if context.GetConfig(ctx).OneShot {
			upload.OneShot = true
		}

		// Set upload TTL
		upload.TTL = context.GetConfig(ctx).DefaultTTL

		// Set upload user and token
		user := context.GetUser(ctx)
		if user != nil {
			upload.User = user.ID
			token := context.GetToken(ctx)
			if token != nil {
				upload.Token = token.Token
			}
		}

		// Save the upload metadata
		err := context.GetMetadataBackend(ctx).CreateUpload(upload)
		if err != nil {
			context.GetLogger(ctx).Warningf("Create new upload error : %s", err)
			context.Fail(ctx, req, resp, "Unable to create new upload", http.StatusInternalServerError)
			return
		}

		// Save upload in the request context
		context.SetUpload(ctx, upload)
		context.SetUploadAdmin(ctx, true)

		// Change the output of the addFile handler
		context.SetQuick(ctx, true)

		next.ServeHTTP(resp, req)
	})
}
