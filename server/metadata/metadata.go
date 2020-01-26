package metadata

import (
	"github.com/root-gg/plik/server/common"
)

// Backend interface describes methods that metadata backends
// must implements to be compatible with plik.
type Backend interface {
	// Create upload metadata
	// TODO : Return nil but no error if upload not found
	CreateUpload(upload *common.Upload) (err error)

	// Get upload metadata
	GetUpload(uploadID string) (upload *common.Upload, err error)

	// Update upload metadata
	UpdateUpload(upload *common.Upload, tx common.UploadTx) (u *common.Upload, err error)

	// Remove upload metadata
	RemoveUpload(upload *common.Upload) (err error)

	// Create user metadata
	CreateUser(user *common.User) (err error)

	// Get user metadata
	// Return nil but no error if user not found
	GetUser(userID string) (user *common.User, err error)

	// Get user metadata from token
	// Return nil but no error if user not found
	GetUserFromToken(token string) (user *common.User, err error)

	// Remove user metadata
	UpdateUser(user *common.User, tx common.UserTx) (u *common.User, err error)

	// Remove user metadata
	RemoveUser(user *common.User) (err error)

	// Get all upload for a given user
	GetUserUploads(user *common.User, token *common.Token) (ids []string, err error)

	// Get statistics for a given user
	GetUserStatistics(user *common.User, token *common.Token) (stats *common.UserStats, err error)

	// Get all users
	GetUsers() (ids []string, err error)

	// Get server statistics
	GetServerStatistics() (stats *common.ServerStats, err error)

	// Return uploads that needs to be removed from the server
	GetUploadsToRemove() (ids []string, err error)
}
