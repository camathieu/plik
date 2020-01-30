package context

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	gocontext "context"
	"github.com/root-gg/juliet"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/metadata"
)

type Context struct {
	config *common.Configuration
	logger *logger.Logger

	metadataBackend metadata.Backend
	dataBackend data.Backend
	streamBackend data.Backend

	sourceIp net.IP

	upload *common.Upload
	file *common.File
	user *common.User
	token *common.Token

	isWhitelisted bool
	isAdmin bool
	isUploadAdmin bool
	isRedirectOnFailure bool
	isQuick bool
	isPanic bool

	req *http.Request
	resp http.ResponseWriter

	goctx gocontext.Context
	mu sync.RWMutex
}

func (ctx *Context) HasConfig() bool {
	ctx.mu.RLock()
	ctx.mu.RUnlock()

	if ctx.config != nil {
		return true
	}
	return false
}

func (ctx *Context) HasLogger() bool {
	ctx.mu.RLock()
	ctx.mu.RUnlock()

	if ctx.logger != nil {
		return true
	}
	return false
}

func (ctx *Context) HasRequest() bool {
	ctx.mu.RLock()
	ctx.mu.RUnlock()

	if ctx.req != nil && ctx.resp != nil {
		return true
	}
	return false
}

// WithConfig set config for the context
func (ctx *Context) WithConfig(config *common.Configuration) *Context {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	if ctx.config == nil {
		ctx.config = config
	} else {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("context configuration overwrite"))
	}
	return ctx
}

// GetConfig from the request context.
func (ctx *Context) GetConfig() (config *common.Configuration) {
	ctx.mu.RLock()
	ctx.mu.RUnlock()

	if ctx.config == nil {
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing context configuration"))
	}
	return ctx.config
}

// WithLogger set config for the context
func (ctx *Context) WithLogger(logger *logger.Logger) *Context {
	ctx.mu.Lock()
	ctx.mu.Unlock()

	if ctx.logger == nil {
		ctx.logger = logger
	} else {
		ctx.isPanic = true
		ctx.InternalServerError(internalServerError, fmt.Errorf("context logger overwrite"))
	}
	return ctx
}

// GetLogger from the request context.
func (ctx *Context) GetLogger() *logger.Logger {
	ctx.mu.RLock()
	ctx.mu.RUnlock()

	if ctx.logger == nil {
		ctx.InternalServerError(internalServerError, fmt.Errorf("missing context logger"))
	}
	return ctx.logger
}

// SetMetadataBackend sets the metadata backend to the context
func SetMetadataBackend(ctx *juliet.Context, backend metadata.Backend) {
	ctx.Set(metadataBackendKey, backend)
}

// GetMetadataBackend from the request context.
func GetMetadataBackend(ctx *juliet.Context) metadata.Backend {
	if backend, ok := ctx.Get(metadataBackendKey); ok {
		return backend.(metadata.Backend)
	}
	return nil
}

// SetDataBackend sets the data backend to the context
func SetDataBackend(ctx *juliet.Context, backend data.Backend) {
	ctx.Set(dataBackendKey, backend)
}

// GetDataBackend from the request context.
func GetDataBackend(ctx *juliet.Context) data.Backend {
	if backend, ok := ctx.Get(dataBackendKey); ok {
		return backend.(data.Backend)
	}
	return nil
}

// SetStreamBackend sets the stream backend to the context
func SetStreamBackend(ctx *juliet.Context, backend data.Backend) {
	ctx.Set(streamBackendKey, backend)
}

// GetStreamBackend from the request context.
func GetStreamBackend(ctx *juliet.Context) data.Backend {
	if backend, ok := ctx.Get(streamBackendKey); ok {
		return backend.(data.Backend)
	}
	return nil
}

// SetSourceIP sets the source ip to the context
func SetSourceIP(ctx *juliet.Context, sourceIP net.IP) {
	ctx.Set(sourceIPKey, sourceIP)
}

// GetSourceIP from the request context.
func GetSourceIP(ctx *juliet.Context) net.IP {
	if sourceIP, ok := ctx.Get(sourceIPKey); ok {
		return sourceIP.(net.IP)
	}
	return nil
}

// SetWhitelisted sets the whitelisted status to the context ( used in tests only )
func SetWhitelisted(ctx *juliet.Context, value bool) {
	ctx.Set(whitelistedKey, value)
}

// IsWhitelisted return true if the IP address in the request context is whitelisted.
func IsWhitelisted(ctx *juliet.Context) bool {
	if whitelisted, ok := ctx.Get(whitelistedKey); ok {
		return whitelisted.(bool)
	}

	uploadWhitelist := GetConfig(ctx).GetUploadWhitelist()

	// Check if the source IP address is in whitelist
	whitelisted := false
	if len(uploadWhitelist) > 0 {
		sourceIP := GetSourceIP(ctx)
		if sourceIP != nil {
			for _, subnet := range uploadWhitelist {
				if subnet.Contains(sourceIP) {
					whitelisted = true
					break
				}
			}
		}
	} else {
		whitelisted = true
	}
	ctx.Set(whitelistedKey, whitelisted)
	return whitelisted
}

// SetUser sets the user to the context
func SetUser(ctx *juliet.Context, user *common.User) {
	ctx.Set(userKey, user)
}

// GetUser from the request context.
func GetUser(ctx *juliet.Context) *common.User {
	if user, ok := ctx.Get(userKey); ok {
		return user.(*common.User)
	}
	return nil
}

// SetToken sets the token to the context
func SetToken(ctx *juliet.Context, token *common.Token) {
	ctx.Set(tokenKey, token)
}

// GetToken from the request context.
func GetToken(ctx *juliet.Context) *common.Token {
	if token, ok := ctx.Get(tokenKey); ok {
		return token.(*common.Token)
	}
	return nil
}

// SetFile sets the file to the context
func SetFile(ctx *juliet.Context, file *common.File) {
	ctx.Set(fileKey, file)
}

// GetFile from the request context.
func GetFile(ctx *juliet.Context) *common.File {
	if file, ok := ctx.Get(fileKey); ok {
		return file.(*common.File)
	}
	return nil
}

// SetUpload sets the upload to the context
func SetUpload(ctx *juliet.Context, upload *common.Upload) {
	ctx.Set(uploadKey, upload)
}

// GetUpload from the request context.
func GetUpload(ctx *juliet.Context) *common.Upload {
	if upload, ok := ctx.Get(uploadKey); ok {
		return upload.(*common.Upload)
	}
	return nil
}

// SetUploadAdmin sets the admin status to the context
func SetUploadAdmin(ctx *juliet.Context, value bool) {
	ctx.Set(uploadAdminKey, value)
}

// IsUploadAdmin returns true if the context has verified that current request can modify the upload
func IsUploadAdmin(ctx *juliet.Context) bool {
	if admin, ok := ctx.Get(uploadAdminKey); ok {
		return admin.(bool)
	}
	return false
}

// SetAdmin sets the upload admin status to the context
func SetAdmin(ctx *juliet.Context, value bool) {
	ctx.Set(adminKey, value)
}

// IsAdmin check if the user is a Plik server administrator
func IsAdmin(ctx *juliet.Context) bool {
	if admin, ok := ctx.Get(adminKey); ok {
		return admin.(bool)
	}
	return false
}

// SetRedirectOnFailure sets the redirect on failure status to the context
func SetRedirectOnFailure(ctx *juliet.Context, value bool) {
	ctx.Set(redirectOnFailureKey, value)
}

// IsRedirectOnFailure return true if the http response should return
// a http redirect instead of an error string.
func IsRedirectOnFailure(ctx *juliet.Context) bool {
	if redirect, ok := ctx.Get(redirectOnFailureKey); ok {
		return redirect.(bool)
	}
	return false
}

// SetQuick sets the quick status to the context
func SetQuick(ctx *juliet.Context, value bool) {
	ctx.Set(quickKey, value)
}

// IsQuick changes the output of the addFile handler
func IsQuick(ctx *juliet.Context) bool {
	if quick, ok := ctx.Get(quickKey); ok {
		return quick.(bool)
	}
	return false
}