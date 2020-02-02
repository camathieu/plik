package plik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipleUploadsInParallel(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	count := 30
	errors := make(chan error, count)
	var wg sync.WaitGroup
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			upload := pc.NewUpload()
			filename := fmt.Sprintf("file_%d", i)
			data := fmt.Sprintf("data data data %s", filename)
			file := upload.AddFileFromReader(filename, bytes.NewBufferString(data))

			err := upload.Upload()
			if err != nil {
				errors <- fmt.Errorf("upload error : %s", err)
				return
			}

			reader, err := file.Download()
			if err != nil {
				errors <- fmt.Errorf("download error : %s", err)
				return
			}
			defer func() { _ = reader.Close() }()

			content, err := ioutil.ReadAll(reader)
			if err != nil {
				errors <- fmt.Errorf("read error : %s", err)
				return
			}
			if string(content) != fmt.Sprintf("data data data %s", file.Name) {
				errors <- fmt.Errorf("file content missmatch")
				return
			}

			err = file.Delete()
			if err != nil {
				errors <- fmt.Errorf("delete error : %s", err)
				return
			}

			_, err = file.Download()
			if err == nil {
				errors <- fmt.Errorf("download deleted file missing error")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		require.NoError(t, err, err.Error())
	}
}

func TestMultipleFilesInParallel(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()

	count := 30
	errors := make(chan error, count)
	var wg sync.WaitGroup
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			filename := fmt.Sprintf("file_%d", i)
			data := fmt.Sprintf("data data data %s", filename)
			file := upload.AddFileFromReader(filename, bytes.NewBufferString(data))

			err := upload.Upload()
			if err != nil {
				errors <- fmt.Errorf("upload error : %s", err)
				return
			}

			reader, err := file.Download()
			if err != nil {
				errors <- fmt.Errorf("download error : %s", err)
				return
			}
			defer func() { _ = reader.Close() }()

			content, err := ioutil.ReadAll(reader)
			if err != nil {
				errors <- fmt.Errorf("read error : %s", err)
				return
			}
			if string(content) != fmt.Sprintf("data data data %s", file.Name) {
				errors <- fmt.Errorf("file content missmatch")
				return
			}

			err = file.Delete()
			if err != nil {
				errors <- fmt.Errorf("delete error : %s", err)
				return
			}

			_, err = file.Download()
			if err == nil {
				errors <- fmt.Errorf("download deleted file missing error")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		require.NoError(t, err, err.Error())
	}
}

func TestUploadDownloadSameFileInParallel(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	filename := fmt.Sprintf("file")
	data := fmt.Sprintf("data data data %s", filename)
	file := upload.AddFileFromReader(filename, bytes.NewBufferString(data))

	count := 30
	errors := make(chan error, count)
	var wg sync.WaitGroup
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := file.Upload()
			if err != nil {
				errors <- fmt.Errorf("upload error : %s", err)
				return
			}

			reader, err := file.Download()
			if err != nil {
				errors <- fmt.Errorf("download error : %s", err)
				return
			}
			defer func() { _ = reader.Close() }()

			content, err := ioutil.ReadAll(reader)
			if err != nil {
				errors <- fmt.Errorf("read error : %s", err)
				return
			}
			if string(content) != fmt.Sprintf("data data data %s", file.Name) {
				errors <- fmt.Errorf("file content missmatch")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		require.NoError(t, err, err.Error())
	}

}