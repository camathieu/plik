package context

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	gocontext "context"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/metadata"
)

type Context struct {
	config              *common.Configuration
	logger              *logger.Logger
	metadataBackend     metadata.Backend
	dataBackend         data.Backend
	streamBackend       data.Backend
	sourceIP            net.IP
	upload              *common.Upload
	file                *common.File
	user                *common.User
	token               *common.Token
	isWhitelisted       bool
	isAdmin             bool
	isUploadAdmin       bool
	isRedirectOnFailure bool
	isQuick             bool
	isPanic             bool
	req                 *http.Request
	resp                http.ResponseWriter
	mu                  sync.RWMutex
}

// SetConfig set config in the context
func (ctx *Context) SetConfig(config *common.Configuration) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.config = config
}

// HasConfig return true if config is set in the context
func (ctx *Context) HasConfig() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.config != nil {
		return true
	}

	return false
}

// GetConfig get config from the context.
func (ctx *Context) GetConfig() *common.Configuration {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.config == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing config from context"))
	}

	return ctx.config
}

// SetLogger set logger in the context
func (ctx *Context) SetLogger(logger *logger.Logger) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.logger = logger
}

// HasLogger return true if logger is set in the context
func (ctx *Context) HasLogger() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.logger != nil {
		return true
	}

	return false
}

// GetLogger get logger from the context.
func (ctx *Context) GetLogger() *logger.Logger {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.logger == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing logger from context"))
	}

	return ctx.logger
}

// SetMetadataBackend set metadataBackend in the context
func (ctx *Context) SetMetadataBackend(metadataBackend metadata.Backend) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.metadataBackend = metadataBackend
}

// HasMetadataBackend return true if metadataBackend is set in the context
func (ctx *Context) HasMetadataBackend() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.metadataBackend != nil {
		return true
	}

	return false
}

// GetMetadataBackend get metadataBackend from the context.
func (ctx *Context) GetMetadataBackend() metadata.Backend {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.metadataBackend == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing metadataBackend from context"))
	}

	return ctx.metadataBackend
}

// SetDataBackend set dataBackend in the context
func (ctx *Context) SetDataBackend(dataBackend data.Backend) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.dataBackend = dataBackend
}

// HasDataBackend return true if dataBackend is set in the context
func (ctx *Context) HasDataBackend() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.dataBackend != nil {
		return true
	}

	return false
}

// GetDataBackend get dataBackend from the context.
func (ctx *Context) GetDataBackend() data.Backend {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.dataBackend == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing dataBackend from context"))
	}

	return ctx.dataBackend
}

// SetStreamBackend set streamBackend in the context
func (ctx *Context) SetStreamBackend(streamBackend data.Backend) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.streamBackend = streamBackend
}

// HasStreamBackend return true if streamBackend is set in the context
func (ctx *Context) HasStreamBackend() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.streamBackend != nil {
		return true
	}

	return false
}

// GetStreamBackend get streamBackend from the context.
func (ctx *Context) GetStreamBackend() data.Backend {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.streamBackend == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing streamBackend from context"))
	}

	return ctx.streamBackend
}

// SetSourceIP set sourceIP in the context
func (ctx *Context) SetSourceIP(sourceIP net.IP) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.sourceIP = sourceIP
}

// HasSourceIP return true if sourceIP is set in the context
func (ctx *Context) HasSourceIP() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.sourceIP != nil {
		return true
	}

	return false
}

// GetSourceIP get sourceIP from the context.
func (ctx *Context) GetSourceIP() net.IP {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.sourceIP == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing sourceIP from context"))
	}

	return ctx.sourceIP
}

// SetUpload set upload in the context
func (ctx *Context) SetUpload(upload *common.Upload) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.upload = upload
}

// HasUpload return true if upload is set in the context
func (ctx *Context) HasUpload() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.upload != nil {
		return true
	}

	return false
}

// GetUpload get upload from the context.
func (ctx *Context) GetUpload() *common.Upload {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.upload == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing upload from context"))
	}

	return ctx.upload
}

// SetFile set file in the context
func (ctx *Context) SetFile(file *common.File) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.file = file
}

// HasFile return true if file is set in the context
func (ctx *Context) HasFile() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.file != nil {
		return true
	}

	return false
}

// GetFile get file from the context.
func (ctx *Context) GetFile() *common.File {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.file == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing file from context"))
	}

	return ctx.file
}

// SetUser set user in the context
func (ctx *Context) SetUser(user *common.User) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.user = user
}

// HasUser return true if user is set in the context
func (ctx *Context) HasUser() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.user != nil {
		return true
	}

	return false
}

// GetUser get user from the context.
func (ctx *Context) GetUser() *common.User {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.user == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing user from context"))
	}

	return ctx.user
}

// SetToken set token in the context
func (ctx *Context) SetToken(token *common.Token) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.token = token
}

// HasToken return true if token is set in the context
func (ctx *Context) HasToken() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.token != nil {
		return true
	}

	return false
}

// GetToken get token from the context.
func (ctx *Context) GetToken() *common.Token {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.token == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing token from context"))
	}

	return ctx.token
}

// SetWhitelisted set isWhitelisted in the context
func (ctx *Context) SetWhitelisted(isWhitelisted bool) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.isWhitelisted = isWhitelisted
}

// IsWhitelisted get isWhitelisted from the context.
func (ctx *Context) IsWhitelisted() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.isWhitelisted
}

// SetAdmin set isAdmin in the context
func (ctx *Context) SetAdmin(isAdmin bool) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.isAdmin = isAdmin
}

// IsAdmin get isAdmin from the context.
func (ctx *Context) IsAdmin() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.isAdmin
}

// SetUploadAdmin set isUploadAdmin in the context
func (ctx *Context) SetUploadAdmin(isUploadAdmin bool) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.isUploadAdmin = isUploadAdmin
}

// IsUploadAdmin get isUploadAdmin from the context.
func (ctx *Context) IsUploadAdmin() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.isUploadAdmin
}

// SetRedirectOnFailure set isRedirectOnFailure in the context
func (ctx *Context) SetRedirectOnFailure(isRedirectOnFailure bool) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.isRedirectOnFailure = isRedirectOnFailure
}

// IsRedirectOnFailure get isRedirectOnFailure from the context.
func (ctx *Context) IsRedirectOnFailure() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.isRedirectOnFailure
}

// SetQuick set isQuick in the context
func (ctx *Context) SetQuick(isQuick bool) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.isQuick = isQuick
}

// IsQuick get isQuick from the context.
func (ctx *Context) IsQuick() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.isQuick
}

// SetReq set req in the context
func (ctx *Context) SetReq(req *http.Request) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.req = req
}

// HasReq return true if req is set in the context
func (ctx *Context) HasReq() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.req != nil {
		return true
	}

	return false
}

// GetReq get req from the context.
func (ctx *Context) GetReq() *http.Request {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.req == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing req from context"))
	}

	return ctx.req
}

// SetResp set resp in the context
func (ctx *Context) SetResp(resp http.ResponseWriter) {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	ctx.resp = resp
}

// HasResp return true if resp is set in the context
func (ctx *Context) HasResp() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.resp != nil {
		return true
	}

	return false
}

// GetResp get resp from the context.
func (ctx *Context) GetResp() http.ResponseWriter {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if ctx.resp == nil {
		ctx.isPanic = true
		ctx.internalServerError(fmt.Errorf("missing resp from context"))
	}

	return ctx.resp
}
