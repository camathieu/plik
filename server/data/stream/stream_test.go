package stream

import (
	"bytes"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func TestAddGetFile(t *testing.T) {
	config := NewConfig(make(map[string]interface{}))
	backend := NewBackend(config)

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		details, err := backend.AddFile(upload, file, bytes.NewBufferString("data"))
		require.NoError(t, err, "unable to add file")
		require.NotNil(t, details, "invalid nil details")
		wg.Done()
	}()

	f := func() {
		for {
			reader, err := backend.GetFile(upload, file)
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			data, err := ioutil.ReadAll(reader)
			require.NoError(t, err, "unable to read reader")

			err = reader.Close()
			require.NoError(t, err, "unable to close reader")

			require.Equal(t, "data", string(data), "invalid reader content")
			break
		}
		wg.Wait()
	}

	err := common.TestTimeout(f, 1*time.Second)
	require.NoError(t, err, "timeout")
}

func TestRemoveFile(t *testing.T) {
	config := NewConfig(make(map[string]interface{}))
	backend := NewBackend(config)

	upload := &common.Upload{}
	upload.Create()
	file := upload.NewFile()

	err := backend.RemoveFile(upload, file)
	require.NoError(t, err)
}
