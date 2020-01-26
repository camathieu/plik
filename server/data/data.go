package data

import (
	"io"

	"github.com/root-gg/plik/server/common"
)

// Backend interface describes methods that data backends
// must implements to be compatible with plik.
type Backend interface {
	GetFile(upload *common.Upload, fileID string) (rc io.ReadCloser, err error)
	AddFile(upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error)
	RemoveFile(upload *common.Upload, fileID string) (err error)
	RemoveUpload(upload *common.Upload) (err error)
}
