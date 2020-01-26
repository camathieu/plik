package testing

import (
	"sync"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/metadata"
)

// Ensure Testing Metadata Backend implements metadata.Backend interface
var _ metadata.Backend = (*Backend)(nil)

// Backend backed in-memory for testing purpose
type Backend struct {
	uploads map[string]*common.Upload
	users   map[string]*common.User

	err error
	mu  sync.Mutex
}

// NewBackend create a new Testing Backend
func NewBackend() (b *Backend) {
	b = new(Backend)
	b.uploads = make(map[string]*common.Upload)
	b.users = make(map[string]*common.User)
	return b
}

// SetError sets the error any subsequent method other call will return
func (b *Backend) SetError(err error) {
	b.err = err
}
