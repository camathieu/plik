package plik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func TestUploadFileTwice(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := &common.Upload{}
	uploadParams, err := pc.create(upload)
	require.NoError(t, err, "unable to create upload")
	require.NotNil(t, uploadParams, "invalid nil uploads params")
	require.NotZero(t, uploadParams.ID, "invalid upload id")

	file := &common.File{}
	file.Name = "filename"

	fileParams, err := pc.uploadFile(uploadParams, file, bytes.NewBufferString("data"))
	require.NoError(t, err, "unable to upload file")
	require.NotNil(t, fileParams, "invalid nil file params")
	require.NotZero(t, fileParams.ID, "invalid file id")

	_, err = pc.uploadFile(uploadParams, fileParams, bytes.NewBufferString("data"))
	require.Error(t, err, "missing error")
	require.Contains(t, err.Error(), "invalid file status uploaded, expected missing", "invalid error")
}

func TestDownloadDuringUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	upload.OneShot = true

	data := "data data data"
	lockedReader := NewLockedReader(bytes.NewBufferString(data))
	file := upload.AddFileFromReader("filename", lockedReader)

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.Metadata().OneShot, "invalid upload non oneshot")

	// The file has not been uploaded
	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : missing", file.Name, file.metadata.ID), "invalid error")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = upload.Upload()
		require.NoError(t, err, "unable to upload file")
		wg.Done()
	}()

	time.Sleep(time.Second)

	// The file is being uploaded
	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : uploading", file.Name, file.metadata.ID), "invalid error")

	lockedReader.Unleash()
	wg.Wait()

	// The file has been uploaded
	reader, err := pc.downloadFile(upload.Metadata(), file.Metadata())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

func TestOneShot(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	pc.OneShot = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", bytes.NewBufferString(data))
	require.NoError(t, err, "unable to upload file")

	require.True(t, upload.Metadata().OneShot, "invalid upload non oneshot")

	reader, err := pc.downloadFile(upload.Metadata(), file.Metadata())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")

	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : removed", file.Name, file.metadata.ID), "invalid error")
}

func TestDownloadOneShotBeforeUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	upload.OneShot = true

	data := "data data data"
	file := upload.AddFileFromReader("filename", bytes.NewBufferString(data))

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.Metadata().OneShot, "invalid upload non oneshot")

	// This should not trigger a file status change and make it impossible to download the file afterwards
	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : missing", file.Name, file.metadata.ID), "invalid error")

	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")

	reader, err := pc.downloadFile(upload.Metadata(), file.Metadata())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")

	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : removed", file.Name, file.metadata.ID), "invalid error")
}

func TestRemoveFileWithoutUploadToken(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", bytes.NewBufferString(data))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Files(), 1, "invalid file count")

	upload.Metadata().UploadToken = ""
	err = pc.removeFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to remove file")
	require.Contains(t, err.Error(), "you are not allowed to remove files from this upload", "invalid error")
}

func TestRemovable(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	pc.Removable = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", bytes.NewBufferString(data))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Files(), 1, "invalid file count")

	upload.Metadata().UploadToken = ""
	err = pc.removeFile(upload.Metadata(), file.Metadata())
	require.NoError(t, err, "unable to upload file")
}

func TestUploadWithoutUploadToken(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.NotNil(t, upload.Metadata(), "upload has not been created")
	require.NotZero(t, upload.ID(), "invalid upload id")

	upload.Metadata().UploadToken = ""
	err = file.Upload()
	require.Error(t, err, "should not be able to upload file to anonymous upload")
	require.Contains(t, err.Error(), "you are not allowed to add file to this upload", "invalid error")
}

func TestStream(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	pc.Stream = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"

	upload := pc.NewUpload()
	file := upload.AddFileFromReader("filename", bytes.NewBufferString(data))

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.Stream, "invalid nil error params")

	errors := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		errors <- upload.Upload()
	}()

	f := func() {
		for {
			time.Sleep(20 * time.Millisecond)
			reader, err := pc.downloadFile(upload.Metadata(), file.Metadata())
			if err != nil {
				continue
			}
			content, err := ioutil.ReadAll(reader)
			require.NoError(t, err, "unable to read file")
			require.Equal(t, data, string(content), "invalid file content")
			break
		}
		wg.Wait()
	}

	err = common.TestTimeout(f, time.Second)
	require.NoError(t, err, "timeout")

	err = <-errors
	require.NoError(t, err, "upload error")

	time.Sleep(time.Second)

	_, err = pc.downloadFile(upload.Metadata(), file.Metadata())
	require.Error(t, err, "unable to download file")
	require.Contains(t, err.Error(), fmt.Sprintf("file %s (%s) is not available : deleted", file.Name, file.metadata.ID), "invalid error")
}

func TestTTL(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	upload.TTL = 1
	err = upload.Create()
	require.NoError(t, err, "unable to upload file")
	require.NotNil(t, upload.Metadata(), "upload has not been created")

	time.Sleep(2 * time.Second)

	_, err = pc.GetUpload(upload.ID())
	require.Error(t, err, "unable to get upload")
	require.Contains(t, err.Error(), "has expired", "upload has not been created")
}
