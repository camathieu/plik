package bolt

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func newBackend(t *testing.T) (backend *Backend, cleanup func()) {
	dir, err := ioutil.TempDir("", "pliktest")
	require.NoError(t, err, "unable to create temp directory")

	backend, err = NewBackend(&Config{Path: dir + "/plik.db"})
	require.NoError(t, err, "unable to create bolt metadata backend")
	cleanup = func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err)
		}
	}

	return backend, cleanup
}

func TestNewConfig(t *testing.T) {
	params := make(map[string]interface{})
	path := "bolt.db"
	params["Path"] = path
	config := NewConfig(params)
	require.Equal(t, path, config.Path)
}

func TestNewBoltMetadataBackend_InvalidPath(t *testing.T) {
	_, err := NewBackend(&Config{Path: string([]byte{0})})
	require.Error(t, err)
}

func TestNewBoltMetadataBackend(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()
	require.NotNil(t, backend)
}
