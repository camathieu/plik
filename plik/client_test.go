/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015> Copyright holders list can be found in AUTHORS file
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package plik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func TestGetServerVersion(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	bi, err := pc.GetServerVersion()
	require.NoError(t, err, "unable to get plik server version")
	require.Equal(t, common.GetBuildInfo().Version, bi.Version, "unable to get plik server version")
}

func TestDefaultUploadParams(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	pc.OneShot = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	require.True(t, upload.OneShot, "upload is not oneshot")

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.HasBeenCreated(), "upload has not been created")
	require.True(t, upload.Info().OneShot, "upload is not oneshot")

	uploadResult, err := pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Equal(t, upload.ID(), uploadResult.ID(), "upload has not been created")
	require.True(t, uploadResult.OneShot, "upload is not oneshot")
}

func TestUploadParamsOverride(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	pc.OneShot = true

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := pc.NewUpload()
	upload.OneShot = false

	err = upload.Create()
	require.NoError(t, err, "unable to create upload")
	require.True(t, upload.HasBeenCreated(), "upload has not been created")
	require.False(t, upload.Info().OneShot, "upload is not oneshot")

	uploadResult, err := pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Equal(t, upload.ID(), uploadResult.ID(), "upload has not been created")
	require.False(t, uploadResult.OneShot, "upload is not oneshot")
}

func TestCreateAndGetUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

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
	defer ps.ShutdownNow()

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
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload, err := pc.UploadReader("filename", bytes.NewBufferString("data"))
	require.NoError(t, err, "unable to create upload")

	uploadToken := upload.Info().UploadToken

	upload, err = pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")

	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))

	upload.Info().UploadToken = uploadToken
	err = upload.Upload()
	fmt.Println(file.Error())
	require.NoError(t, err, "unable to upload file")
	require.NoError(t, file.Error(), "invalid file error")
	require.True(t, file.HasBeenUploaded(), "invalid file has been uploaded status")

	upload, err = pc.GetUpload(upload.ID())
	require.NoError(t, err, "unable to get upload")
	require.Len(t, upload.files, 2, "invalid file count")
}

func TestUploadReader(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 1, "invalid file count")

	file := upload.Files()[0]
	reader, err := pc.DownloadFile(upload.Info(), file.Info())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

func TestUploadReadCloser(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 1, "invalid file count")

	file := upload.Files()[0]
	reader, err := pc.DownloadFile(upload.Info(), file.Info())
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

func TestUploadFiles(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

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

	upload, err := pc.UploadFiles(tmpFile.Name(), tmpFile2.Name())
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 2, "invalid file count")

	for _, file := range upload.Info().Files {
		reader, err := pc.DownloadFile(upload.Info(), file)
		require.NoError(t, err, "unable to download file")
		content, err := ioutil.ReadAll(reader)
		require.NoError(t, err, "unable to read file")
		require.Equal(t, data, string(content), "invalid file content")
	}
}

func TestUploadFileNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	_, err = pc.UploadFiles("missing_file_name")
	require.Error(t, err, "unable to upload file")
	require.Contains(t, err.Error(), "not found", "unable to upload file")

	_, err = pc.UploadFiles(".")
	require.Error(t, err, "unable to upload file")
	require.Contains(t, err.Error(), "Unhandled file mode", "unable to upload file")
}

func TestRemoveFile(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 1, "invalid file count")

	file := upload.Files()[0]
	_, err = pc.DownloadFile(upload.Info(), file.Info())
	require.NoError(t, err, "unable to download file")

	err = pc.RemoveFile(upload.Info(), file.Info())
	require.NoError(t, err, "unable to remove file")

	_, err = pc.DownloadFile(upload.Info(), file.Info())
	common.RequireError(t, err, "not found")
}

func TestRemoveFileNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	file := &common.File{}
	file.ID = "blah"
	err = pc.RemoveFile(upload, file)
	common.RequireError(t, err, "not found")
}

func TestRemoveFileNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	upload := common.NewUpload()
	upload.Create()
	file := &common.File{}
	file.ID = "blah"
	err := pc.RemoveFile(upload, file)
	common.RequireError(t, err, "connection refused")
}

func TestRemoveUpload(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 1, "invalid file count")

	file := upload.Files()[0]
	_, err = pc.DownloadFile(upload.Info(), file.Info())
	require.NoError(t, err, "unable to download file")

	err = pc.RemoveUpload(upload.Info())
	require.NoError(t, err, "unable to remove upload")

	_, err = pc.GetUpload(upload.ID())
	common.RequireError(t, err, "not found")

	_, err = pc.DownloadFile(upload.Info(), file.Info())
	common.RequireError(t, err, "not found")
}

func TestRemoveUploadNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	err = pc.RemoveUpload(upload)
	common.RequireError(t, err, "not found")
}

func TestRemoveUploadNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	upload := common.NewUpload()
	upload.Create()
	err := pc.RemoveUpload(upload)
	common.RequireError(t, err, "connection refused")
}

func TestGetArchive(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Info().Files, 1, "invalid file count")

	reader, err := pc.DownloadArchive(upload.Info())
	require.NoError(t, err, "unable to download archive")

	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read archive")

	require.NotEmpty(t, content, "empty archive")
}

func TestGetArchiveNotFound(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := start(ps)
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	_, err = pc.DownloadArchive(upload)
	common.RequireError(t, err, "not found")
}

func TestGetArchiveNoServer(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	upload := common.NewUpload()
	upload.Create()
	_, err := pc.DownloadArchive(upload)
	common.RequireError(t, err, "connection refused")
}
