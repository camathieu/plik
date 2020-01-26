package plik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func TestGetServerVersion(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	bi, err := pc.GetServerVersion()
	require.NoError(t, err, "unable to get plik server version")
	require.Equal(t, common.GetBuildInfo().Version, bi.Version, "unable to get plik server version")
}

func TestDefaultUploadParams(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	pc.OneShot = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	require.True(t, upload.OneShot, "upload is not oneshot")

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.HasBeenCreated(), "upload has not been created")
	require.True(t, upload.details.OneShot, "upload is not oneshot")

	uploadResult, err := pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Equal(t, upload.ID(), uploadResult.ID(), "upload has not been created")
	require.True(t, uploadResult.OneShot, "upload is not oneshot")
}

func TestUploadParamsOverride(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	pc.OneShot = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	upload.OneShot = false

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.HasBeenCreated(), "upload has not been created")
	require.False(t, upload.Details().OneShot, "upload is not oneshot")

	uploadResult, err := pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Equal(t, upload.ID(), uploadResult.ID(), "upload has not been created")
	require.False(t, uploadResult.OneShot, "upload is not oneshot")
}

func TestCreateAndGetUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	err = upload.Create()
	require.NoError(t, err, "unable to upload file")
	require.True(t, upload.HasBeenCreated(), "upload has not been created")

	uploadResult, err := pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to upload file")
	require.Equal(t, upload.ID(), uploadResult.ID(), "upload has not been created")
}

func TestAddFileToExistingUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	err = upload.Create()
	require.NoError(t, err, "unable to create upload")

	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))

	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")
	require.NoError(t, file.Error(), "invalid file error")
	require.True(t, file.HasBeenUploaded(), "invalid file has been uploaded status")
}

func TestAddFileToExistingUpload2(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload, _, err := pc.UploadReader("filename", bytes.NewBufferString("data"))
	require.NoError(t, err, "unable to create upload")

	uploadToken := upload.details.UploadToken

	upload, err = pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")

	file2 := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))

	upload.details.UploadToken = uploadToken
	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")
	require.NoError(t, file2.Error(), "invalid file error")
	require.True(t, file2.HasBeenUploaded(), "invalid file has been uploaded status")

	upload, err = pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Len(t, upload.files, 2, "invalid file count")
}

func TestUploadReader(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 1, "invalid file count")

	reader, err := pc.downloadFile(upload.getParams(), file.getParams())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

func TestUploadReadCloser(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 1, "invalid file count")

	reader, err := pc.downloadFile(upload.getParams(), file.getParams())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

func TestUploadFiles(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	tmpFile, err := ioutil.TempFile("", "pliktmpfile")
	require.NoError(t, err, "unable to create tmp file")
	defer os.Remove(tmpFile.Name())

	data := "data data data"
	_, err = tmpFile.Write([]byte(data))
	require.NoError(t, err, "unable to write tmp file")
	err = tmpFile.Close()
	require.NoError(t, err, "unable to close tmp file")

	tmpFile2, err := ioutil.TempFile("", "pliktmpfile")
	require.NoError(t, err, "unable to create tmp file")
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile2.Write([]byte(data))
	require.NoError(t, err, "unable to write tmp file")
	err = tmpFile2.Close()
	require.NoError(t, err, "unable to close tmp file")

	upload := pc.NewUpload()

	_, err = upload.AddFileFromPath(tmpFile.Name())
	require.NoError(t, err, "unable to add file")

	_, err = upload.AddFileFromPath(tmpFile2.Name())
	require.NoError(t, err, "unable to add file")

	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 2, "invalid file count")

	for _, file := range upload.Details().Files {
		reader, err := pc.downloadFile(upload.Details(), file)
		require.NoError(t, err, "unable to download file")
		content, err := ioutil.ReadAll(reader)
		require.NoError(t, err, "unable to read file")
		require.Equal(t, data, string(content), "invalid file content")
	}
}

func TestUploadMultipleFiles(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	for i := 1; i <= 30; i++ {
		filename := fmt.Sprintf("file_%d", i)
		data := fmt.Sprintf("data data data %s", filename)
		upload.AddFileFromReader(filename, bytes.NewBufferString(data))
	}

	err = upload.Upload()
	require.NoError(t, err, "unable to upload files")

	for _, file := range upload.Files() {
		require.True(t, file.HasBeenUploaded(), "file has not been uploaded")
		require.NoError(t, file.Error(), "unexpected file error")

		reader, err := pc.downloadFile(upload.Details(), file.Details())
		require.NoError(t, err, "unable to download file")
		content, err := ioutil.ReadAll(reader)
		require.NoError(t, err, "unable to read file")
		require.Equal(t, fmt.Sprintf("data data data %s", file.Name), string(content), "invalid file content")
	}
}

func TestMultipleUploadsInParallel(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func(i int) {
			upload := pc.NewUpload()

			filename := fmt.Sprintf("file_%d", i)
			data := fmt.Sprintf("data data data %s", filename)
			file := upload.AddFileFromReader(filename, bytes.NewBufferString(data))

			err := upload.Upload()
			require.NoError(t, err, "unable to upload file")

			reader, err := file.Download()
			require.NoError(t, err, "unable to download file")

			content, err := ioutil.ReadAll(reader)
			require.NoError(t, err, "unable to read file")
			require.Equal(t, fmt.Sprintf("data data data %s", file.Name), string(content), "invalid file content")

			err = file.Delete()
			require.NoError(t, err, "unable to delete file")

			_, err = file.Download()
			require.Error(t, err, "missing expected error")

			wg.Done()
		}(i)
	}

	wg.Wait()
}

// TODO This is full of races !!!
//func TestMultipleFilesInParallel(t *testing.T) {
//	ps, pc := newPlikServerAndClient()
//	defer shutdown(ps)
//
//	err := start(ps)
//	require.NoError(t, err, "unable to start plik server")
//
//	upload := pc.NewUpload()
//
//	var wg sync.WaitGroup
//	for i := 1; i <= 30; i++ {
//		wg.Add(1)
//		go func(i int) {
//			filename := fmt.Sprintf("file_%d", i)
//			data := fmt.Sprintf("data data data %s", filename)
//			file := upload.AddFileFromReader(filename, bytes.NewBufferString(data))
//
//			err := upload.Upload()
//			require.NoError(t, err, "unable to upload file")
//
//			reader, err := file.Download()
//			require.NoError(t, err, "unable to download file")
//
//			content, err := ioutil.ReadAll(reader)
//			require.NoError(t, err, "unable to read file")
//			require.Equal(t, fmt.Sprintf("data data data %s", file.Name), string(content), "invalid file content")
//
//			err = file.Delete()
//			require.NoError(t, err, "unable to delete file")
//
//			_, err = file.Download()
//			require.Error(t, err, "missing expected error")
//
//			wg.Done()
//		}(i)
//	}
//
//	wg.Wait()
//}

func TestUploadFileNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	_, _, err = pc.UploadFile("missing_file_name")
	require.Error(t, err, "unable to upload file")
	require.Contains(t, err.Error(), "not found", "unable to upload file")

	_, _, err = pc.UploadFile(".")
	require.Error(t, err, "unable to upload file")
	require.Contains(t, err.Error(), "Unhandled file mode", "unable to upload file")
}

func TestRemoveFile(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 1, "invalid file count")

	_, err = pc.downloadFile(upload.getParams(), file.getParams())
	require.NoError(t, err, "unable to download file")

	err = pc.removeFile(upload.Details(), file.Details())
	require.NoError(t, err, "unable to remove file")

	_, err = pc.downloadFile(upload.getParams(), file.getParams())
	common.RequireError(t, err, "not found")
}

func TestRemoveFileNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	file := &common.File{}
	file.ID = "blah"
	err = pc.removeFile(upload, file)
	common.RequireError(t, err, "not found")
}

func TestRemoveFileNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	upload := common.NewUpload()
	upload.Create()
	file := &common.File{}
	file.ID = "blah"
	err := pc.removeFile(upload, file)
	common.RequireError(t, err, "connection refused")
}

func TestDeleteUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 1, "invalid file count")

	_, err = file.Download()
	require.NoError(t, err, "unable to download file")

	err = upload.Delete()
	require.NoError(t, err, "unable to remove upload")

	_, err = pc.GetUpload(upload.ID())
	common.RequireError(t, err, "not found")

	_, err = file.Download()
	common.RequireError(t, err, "not found")
}

func TestDeleteUploadNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	err = pc.removeUpload(upload)
	common.RequireError(t, err, "not found")

	upload2 := pc.NewUpload()
	err = upload2.Delete()
	common.RequireError(t, err, "not found")
}

func TestDeleteUploadNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	upload := common.NewUpload()
	upload.Create()
	err := pc.removeUpload(upload)
	common.RequireError(t, err, "connection refused")
}

func TestDownloadArchive(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, _, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Details().Files, 1, "invalid file count")

	reader, err := upload.DownloadZipArchive()
	require.NoError(t, err, "unable to download archive")

	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read archive")

	require.NotEmpty(t, content, "empty archive")
}

func TestGetArchiveNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	_, err = pc.downloadArchive(upload)
	common.RequireError(t, err, "not found")

	upload2 := pc.NewUpload()
	_, err = upload2.DownloadZipArchive()
	common.RequireError(t, err, "not found")
}

func TestGetArchiveNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	upload := common.NewUpload()
	upload.Create()
	_, err := pc.downloadArchive(upload)
	common.RequireError(t, err, "connection refused")
}