package data

import (
	"io"

	"github.com/root-gg/plik/server/common"
)

// Backend interface describes methods that data backends
// must implements to be compatible with plik.
type Backend interface {
	AddFile(upload *common.Upload, file *common.File, reader io.Reader) (backendDetails string, err error)
	GetFile(upload *common.Upload, file *common.File) (reader io.ReadCloser, err error)
	RemoveFile(upload *common.Upload, file *common.File) (err error)
}
