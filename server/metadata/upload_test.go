package metadata

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createUpload(t *testing.T, b *Backend, upload *common.Upload) {
	upload.PrepareInsertForTests()
	err := b.CreateUpload(upload)
	require.NoError(t, err, "create upload error : %s", err)
}

func TestBackend_CreateUpload(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	file := upload.NewFile()

	createUpload(t, b, upload)

	require.NotZero(t, upload.ID, "missing upload id")
	require.NotZero(t, upload.CreatedAt, "missing creation date")
	require.NotZero(t, file.ID, "missing file id")
	require.Equal(t, upload.ID, file.UploadID, "missing file id")
	require.NotZero(t, file.CreatedAt, "missing creation date")
}

func TestBackend_GetUpload(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	_ = upload.NewFile()

	createUpload(t, b, upload)

	result, err := b.GetUpload(upload.ID)
	require.NoError(t, err, "get upload error")

	require.Equal(t, upload.ID, result.ID, "invalid upload id")
	require.Zero(t, result.Files, "invalid upload files")
	require.Equal(t, upload.UploadToken, result.UploadToken, "invalid upload token")
}

func TestBackend_GetUpload_NotFound(t *testing.T) {
	b := newTestMetadataBackend()

	upload, err := b.GetUpload("not found")
	require.NoError(t, err, "get upload error")
	require.Nil(t, upload, "upload not nil")
}

func TestBackend_DeleteUpload(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	_ = upload.NewFile()

	createUpload(t, b, upload)

	err := b.DeleteUpload(upload)
	require.NoError(t, err, "get upload error")

	upload, err = b.GetUpload(upload.ID)
	require.NoError(t, err, "get upload error")
	require.Nil(t, upload, "upload not nil")
}

func TestBackend_CountUploadFiles(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	_ = upload.NewFile()

	createUpload(t, b, upload)

	count, err := b.CountUploadFiles(upload)
	require.NoError(t, err, "count upload files error")
	require.Equal(t, 1, count, "count upload files mismatch")
}

//
//func TestBackend_DeleteUploadWithFiles(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	file := upload.NewFile()
//
//    createUpload(t, b, upload)
//
//	content := bytes.NewBufferString("data data data")
//	_, err := b.dataBackend.AddFile(upload, file, content)
//	require.NoError(t, err, "add file error")
//
//	err = b.UpdateFileStatus(file, common.FileMissing, common.FileUploaded)
//	require.NoError(t, err, "update file status error")
//
//	err = b.DeleteUpload(upload)
//	require.NoError(t, err, "delete upload error")
//
//	upload, err = b.GetUpload(upload.ID)
//	require.NoError(t, err, "get upload error")
//	require.Nil(t, upload, "upload not nil")
//
//	_, err = b.dataBackend.GetFile(upload, file)
//	require.Error(t, err, "get file error")
//}
//
func TestBackend_DeleteExpiredUploads(t *testing.T) {
	b := newTestMetadataBackend()

	upload1 := &common.Upload{}
	createUpload(t, b, upload1)

	upload2 := &common.Upload{}
	createUpload(t, b, upload2)

	deadline2 := time.Now().Add(time.Hour)
	upload2.ExpireAt = &deadline2
	err := b.db.Save(upload2).Error
	require.NoError(t, err, "update upload error")

	upload3 := &common.Upload{}
	createUpload(t, b, upload3)

	deadline3 := time.Now().Add(-time.Hour)
	upload3.ExpireAt = &deadline3
	err = b.db.Save(upload3).Error
	require.NoError(t, err, "update upload error")

	removed, err := b.DeleteExpiredUploads()
	require.Nil(t, err, "delete expired upload error")
	require.Equal(t, 1, removed, "removed expired upload count mismatch")
}
