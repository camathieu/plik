package stream

import (
	"bytes"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/root-gg/juliet"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func newTestingContext(config *common.Configuration) (ctx *juliet.Context) {
	ctx = juliet.NewContext()
	context.SetConfig(ctx, config)
	context.SetLogger(ctx, logger.NewLogger())
	return ctx
}

func TestAddGetFile(t *testing.T) {
	config := NewConfig(make(map[string]interface{}))
	backend := NewBackend(config)

	upload := common.NewUpload()
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
			reader, err := backend.GetFile(upload, file.ID)
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

	upload := common.NewUpload()
	upload.Create()
	file := upload.NewFile()

	err := backend.RemoveFile(upload, file.ID)
	require.Error(t, err, "able to remove file")
}

func TestRemoveUpload(t *testing.T) {
	config := NewConfig(make(map[string]interface{}))
	backend := NewBackend(config)

	upload := common.NewUpload()
	upload.Create()

	err := backend.RemoveUpload(upload)
	require.Error(t, err, "able to remove upload")
}
