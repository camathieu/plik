package metadata

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBackend_CreateFile(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}

	createUpload(t, b, upload)

	file := upload.NewFile()
	err := b.CreateFile(file)
	require.NoError(t, err, "create file error")
}

func TestBackend_CreateFile_UploadNotFound(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	upload.ID = "nope"

	file := upload.NewFile()
	file.GenerateID()

	err := b.CreateFile(file)
	require.Error(t, err, "no create file error")
}

func TestBackend_GetFile(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	file := upload.NewFile()

	createUpload(t, b, upload)

	result, err := b.GetFile(file.ID)
	require.NoError(t, err, "create file error")

	require.NotNil(t, file, "missing file")
	require.Equal(t, file.ID, result.ID, "invalid file id")
}

func TestBackend_GetFile_NotFound(t *testing.T) {
	b := newTestMetadataBackend()

	file, err := b.GetFile("not found")
	require.NoError(t, err, "get file error")
	require.Nil(t, file, "file not nil")
}

func TestBackend_GetFiles(t *testing.T) {
	b := newTestMetadataBackend()

	// To spice the test
	upload := &common.Upload{}
	_ = upload.NewFile()
	createUpload(t, b, upload)

	upload = &common.Upload{}
	_ = upload.NewFile()
	_ = upload.NewFile()
	createUpload(t, b, upload)

	files, err := b.GetFiles(upload)
	require.NoError(t, err, "create file error")
	require.Len(t, files, 2, "missing files")
}

func TestBackend_UpdateFile(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	file := upload.NewFile()

	createUpload(t, b, upload)

	file.Status = common.FileUploaded
	file.Name = "name"
	file.Md5 = "md5"
	err := b.UpdateFile(file, common.FileMissing)
	require.NoError(t, err, "create file error")

	result, err := b.GetFile(file.ID)
	require.NoError(t, err, "create file error")

	require.NotNil(t, file, "missing file")
	require.Equal(t, file.ID, result.ID, "invalid file id")
	require.Equal(t, file.Name, result.Name, "invalid file name")
	require.Equal(t, file.Md5, result.Md5, "invalid file md5")
	require.Equal(t, file.Status, result.Status, "invalid file md5")
}

//func TestBackend_DeleteFile(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	file := upload.NewFile()
//
//    createUpload(t, b, upload)
//
//	fileID := file.ID
//	content := bytes.NewBufferString("data data data")
//	_, err := b.dataBackend.AddFile(upload, file, content)
//	require.NoError(t, err, "add file error")
//
//	err = b.UpdateFileStatus(file, common.FileMissing, common.FileUploaded)
//	require.NoError(t, err, "update file status error")
//
//	err = b.DeleteFile(upload, file)
//	require.NoError(t, err, "delete file error")
//
//	file, err = b.GetFile(fileID)
//	require.NoError(t, err, "get file error")
//	require.Equal(t, common.FileDeleted, file.Status, "file not nil")
//
//	_, err = b.dataBackend.GetFile(upload, file)
//	require.Error(t, err, "get file error")
//}
//
//func TestBackend_DeleteFile_NotFound(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	file := upload.NewFile()
//	file.GenerateID()
//
//	file.Status = common.FileRemoved
//	err := b.DeleteFile(upload, file)
//	require.Error(t, err, "no delete file error")
//}
//
//func TestBackend_DeleteFile_Missing(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	file := upload.NewFile()
//
//    createUpload(t, b, upload)
//
//	fileID := file.ID
//	content := bytes.NewBufferString("data data data")
//	_, err := b.dataBackend.AddFile(upload, file, content)
//	require.NoError(t, err, "add file error")
//
//	err = b.DeleteFile(upload, file)
//	require.NoError(t, err, "delete file error")
//
//	file, err = b.GetFile(fileID)
//	require.NoError(t, err, "get file error")
//	require.Equal(t, common.FileDeleted, file.Status, "file not deleted")
//
//	// Check deleted also
//	err = b.DeleteFile(upload, file)
//	require.NoError(t, err, "delete file error")
//
//	file, err = b.GetFile(fileID)
//	require.NoError(t, err, "get file error")
//	require.Equal(t, common.FileDeleted, file.Status, "file not deleted")
//}
//
//func TestBackend_DeleteFile_Removed(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	file := upload.NewFile()
//
//    createUpload(t, b, upload)
//
//	fileID := file.ID
//	content := bytes.NewBufferString("data data data")
//	_, err := b.dataBackend.AddFile(upload, file, content)
//	require.NoError(t, err, "add file error")
//
//	err = b.UpdateFileStatus(file, common.FileMissing, common.FileRemoved)
//	require.NoError(t, err, "update file status error")
//
//	err = b.DeleteFile(upload, file)
//	require.NoError(t, err, "delete file error")
//
//	file, err = b.GetFile(fileID)
//	require.NoError(t, err, "get file error")
//	require.Equal(t, common.FileDeleted, file.Status, "file not deleted")
//}
//
//func TestBackend_ForEachUploadFiles(t *testing.T) {
//	b := newTestMetadataBackend()
//
//	upload := &common.Upload{}
//	_ = upload.NewFile()
//	_ = upload.NewFile()
//
//    createUpload(t, b, upload)
//
//	var files []*common.File
//	f := func(file *common.File) error {
//		files = append(files, file)
//		return nil
//	}
//
//	err := b.ForEachUploadFiles(upload, f)
//	require.NoError(t, err, "for each upload file error")
//	require.Len(t, files, 2, "file count mismatch")
//
//}
