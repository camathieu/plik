package metadata

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/utils"
	"github.com/stretchr/testify/require"
)

func newTestMetadataBackend() *Backend {
	config := &Config{Driver: "sqlite3", ConnectionString: "/tmp/plik.test.db", EraseFirst: true}

	b, err := NewBackend(config)
	if err != nil {
		panic("unable to create metadata backend")
	}

	return b
}

func TestMetadata(t *testing.T) {
	b := newTestMetadataBackend()

	uploadID := "azertiop"
	upload := &common.Upload{ID: uploadID}

	err := b.db.Create(upload).Error
	require.NoError(t, err, "unable to create upload")

	file := &common.File{ID: "1234567890", UploadID: uploadID}
	upload.Files = append(upload.Files, file)

	err = b.db.Save(&upload).Error
	require.NoError(t, err, "unable to update upload")

	upload = &common.Upload{}
	err = b.db.Preload("Files").Take(upload, "id = ?", uploadID).Error
	require.NoError(t, err, "unable to fetch upload")
}

func TestGormConcurrent(t *testing.T) {
	type Object struct {
		gorm.Model
		Foo string
	}

	// https://github.com/jinzhu/gorm/issues/2875
	db, err := gorm.Open("sqlite3", "/tmp/plik.db")
	require.NoError(t, err, "DB open error")

	err = db.AutoMigrate(&Object{}).Error
	require.NoError(t, err, "schema update error")

	count := 30
	var wg sync.WaitGroup
	errors := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			errors <- db.Create(&Object{Foo: fmt.Sprintf("%d", i)}).Error
		}(i)
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err, "unexpected error")
	}
}

func TestMetadataConcurrent(t *testing.T) {
	b := newTestMetadataBackend()

	uploadID := "azertiop"
	upload := &common.Upload{ID: uploadID}

	err := b.db.Create(upload).Error
	require.NoError(t, err, "unable to create upload")

	count := 30
	var wg sync.WaitGroup
	errors := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			errors <- b.db.Create(&common.File{ID: fmt.Sprintf("file_%d", i), UploadID: uploadID}).Error
		}(i)
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err, "unexpected error")
	}

	upload = &common.Upload{}
	err = b.db.Preload("Files").Take(upload, "id = ?", uploadID).Error
	require.NoError(t, err, "unable to fetch upload")

	utils.Dump(upload)
}

func TestMetadataUpdateFileStatus(t *testing.T) {
	b := newTestMetadataBackend()

	uploadID := "azertiop"
	upload := &common.Upload{ID: uploadID}

	err := b.db.Create(upload).Error
	require.NoError(t, err, "unable to create upload")

	file := &common.File{ID: "1234567890", UploadID: uploadID, Status: common.FileMissing}
	upload.Files = append(upload.Files, file)

	err = b.db.Save(&upload).Error
	require.NoError(t, err, "unable to update upload")

	file.Status = common.FileUploaded
	result := b.db.Where(&common.File{Status: common.FileUploading}).Save(&file)
	require.NoError(t, result.Error, "unable to update missing file")
	require.Equal(t, int64(0), result.RowsAffected, "unexpected update")

	result = b.db.Where(&common.File{Status: common.FileMissing}).Save(&file)
	require.NoError(t, result.Error, "unable to update missing file")
	require.Equal(t, int64(1), result.RowsAffected, "unexpected update")

	upload = &common.Upload{}
	err = b.db.Preload("Files").Take(upload, "id = ?", uploadID).Error
	require.NoError(t, err, "unable to fetch upload")

	utils.Dump(upload)
}

func TestMetadataNotFound(t *testing.T) {
	b := newTestMetadataBackend()

	upload := &common.Upload{}
	err := b.db.Where(&common.Upload{ID: "notfound"}).Take(upload).Error
	require.Error(t, err, "unable to fetch upload")
	require.True(t, gorm.IsRecordNotFoundError(err), "unexpected error type")

	utils.Dump(upload)
}

func TestMetadataCursor(t *testing.T) {
	b := newTestMetadataBackend()

	var expected = []string{"upload 1", "upload 2", "upload 3"}
	for _, id := range expected {
		err := b.db.Create(&common.Upload{ID: id}).Error
		require.NoError(t, err, "unable to create upload")
	}

	rows, err := b.db.Model(&common.Upload{}).Rows()
	require.NoError(t, err, "unable to fetch uploads")

	var ids []string
	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		require.NoError(t, err, "unable to read row")
		ids = append(ids, upload.ID)
	}

	require.Equal(t, expected, ids, "mismatch")
}

func TestMetadataExpiredCursor(t *testing.T) {
	b := newTestMetadataBackend()

	err := b.db.Create(&common.Upload{ID: "upload 1"}).Error
	require.NoError(t, err, "unable to create upload")

	expire := time.Now().Add(time.Hour)
	err = b.db.Create(&common.Upload{ID: "upload 2", ExpireAt: &expire}).Error
	require.NoError(t, err, "unable to create upload")

	expire2 := time.Now().Add(-time.Hour)
	err = b.db.Create(&common.Upload{ID: "upload 3", ExpireAt: &expire2}).Error
	require.NoError(t, err, "unable to create upload")

	rows, err := b.db.Model(&common.Upload{}).Where("expire_at < ?", time.Now()).Rows()
	require.NoError(t, err, "unable to fetch uploads")

	var ids []string
	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		require.NoError(t, err, "unable to read row")
		ids = append(ids, upload.ID)
	}

	require.Equal(t, []string{"upload 3"}, ids, "mismatch")
}

// https://github.com/mattn/go-sqlite3/issues/569
func TestMetadataCursorLock(t *testing.T) {
	b := newTestMetadataBackend()

	var expected = []string{"upload 1", "upload 2", "upload 3"}
	for _, id := range expected {
		err := b.db.Create(&common.Upload{ID: id}).Error
		require.NoError(t, err, "unable to create upload")
	}

	rows, err := b.db.Model(&common.Upload{}).Rows()
	require.NoError(t, err, "unable to select uploads")

	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		require.NoError(t, err, "unable to read row")

		upload.Comments = "lol"
		err = b.db.Save(upload).Error
		require.NoError(t, err, "unable to save upload")
	}
}

func TestUnscoped(t *testing.T) {
	b := newTestMetadataBackend()

	uploadID := "azertiop"
	upload := &common.Upload{ID: uploadID}

	err := b.db.Create(upload).Error
	require.NoError(t, err, "unable to create upload")

	var count int
	err = b.db.Model(&common.Upload{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count).Error
	require.NoError(t, err, "get deleted upload error")
	require.Equal(t, 0, count, "get deleted upload count error")

	err = b.db.Delete(&upload).Error
	require.NoError(t, err, "unable to delete upload")

	upload = &common.Upload{}
	err = b.db.Take(upload, &common.Upload{ID: uploadID}).Error
	require.True(t, gorm.IsRecordNotFoundError(err), "get upload error")

	upload = &common.Upload{}
	err = b.db.Unscoped().Take(upload, &common.Upload{ID: uploadID}).Error
	require.NoError(t, err, "get upload error")
	require.NotNil(t, upload, "get upload nil")

	err = b.db.Model(&common.Upload{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count).Error
	require.NoError(t, err, "get deleted upload error")
	require.Equal(t, 1, count, "get deleted upload count error")
}

func TestDoubleDelete(t *testing.T) {
	b := newTestMetadataBackend()

	uploadID := "azertiop"
	upload := &common.Upload{ID: uploadID}

	err := b.db.Create(upload).Error
	require.NoError(t, err, "unable to create upload")

	err = b.db.Delete(&upload).Error
	require.NoError(t, err, "unable to delete upload")

	err = b.db.Delete(&upload).Error
	require.NoError(t, err, "unable to delete upload")
}
