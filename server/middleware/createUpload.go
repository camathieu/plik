package middleware

import (
	"fmt"
	"net/http"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// CreateUpload create a new upload on the fly to be used in the next handler
func CreateUpload(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		upload := common.NewUpload()
		upload.Create()

		// Set upload remote IP
		upload.RemoteIP = ctx.GetSourceIP().String()

		// Set upload one shot mode
		if ctx.GetConfig().OneShot {
			upload.OneShot = true
		}

		// Set upload TTL
		upload.TTL = ctx.GetConfig().DefaultTTL

		// Set upload user and token
		user := ctx.GetUser()
		if user != nil {
			upload.User = user.ID
			token := ctx.GetToken()
			if token != nil {
				upload.Token = token.Token
			}
		}

		// Save the upload metadata
		err := ctx.GetMetadataBackend().CreateUpload(upload)
		if err != nil {
			ctx.InternalServerError(fmt.Errorf("unable to create upload : %s", err))
			return
		}

		// Save upload in the request context
		ctx.SetUpload(upload)
		ctx.SetUploadAdmin(true)

		// Change the output of the addFile handler
		ctx.SetQuick(true)

		next.ServeHTTP(resp, req)
	})
}
