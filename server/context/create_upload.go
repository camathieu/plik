package context

import "github.com/root-gg/plik/server/common"

func (ctx *Context) SetUploadContext(upload *common.Upload) {
	if ctx.GetSourceIP() != nil {
		// Set upload remote IP
		upload.RemoteIP = ctx.GetSourceIP().String()
	}

	// Set upload user and token
	user := ctx.GetUser()
	if user != nil {
		upload.User = user.ID
		token := ctx.GetToken()
		if token != nil {
			token := ctx.GetToken()
			if token != nil {
				upload.Token = token.Token
			}
		}
	}
}