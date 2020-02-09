package metadata

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
	"time"
)

func (b *Backend) CreateUpload(upload *common.Upload) (err error) {
	return b.db.Create(upload).Error
}

func (b *Backend) GetUpload(ID string) (upload *common.Upload, err error) {
	upload = &common.Upload{}
	err = b.db.Where(&common.Upload{ID: ID}).Take(upload).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return upload, err
}

func (b *Backend) DeleteUpload(upload *common.Upload) error {

	// Try to remove as much files as possible
	var errors []error
	for _, file := range upload.Files {
		err := b.DeleteFile(upload, file)
		if err != nil {
			b.log.Warningf("unable to delete file %s : %s", file.ID, err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("unable to delete %d files", len(errors))
	}

	// Delete upload
	err := b.db.Transaction(func(tx *gorm.DB) (err error) {
		// Check upload files status
		var files []*common.File
		result := b.db.Not(&common.File{UploadID: upload.ID, Status: common.FileDeleted}).Find(&files)
		if result.Error != nil {
			return fmt.Errorf("unable to get upload files : %s", result.Error)
		}
		if len(files) > 1 {
			return fmt.Errorf("not all upload files have been deleted")
		}

		// Delete upload files
		err = tx.Delete(&common.File{UploadID: upload.ID}).Error
		if err != nil {
			return fmt.Errorf("unable to delete files metadata")
		}

		// Delete upload
		err = tx.Unscoped().Delete(upload).Error
		if err != nil {
			return fmt.Errorf("unable to delete upload metadata")
		}

		return nil
	})

	return err
}

func (b *Backend) DeleteExpiredUploads() (removed int, err error) {
	rows, err := b.db.Model(&common.Upload{}).Where("expired_at > ?", time.Now()).Rows()
	defer func() { _ = rows.Close() }()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch expired upload : %s", err)
	}

	var errors []error
	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		if err != nil {
			return 0, fmt.Errorf("unable to fetch next expired upload : %s", err)
		}

		err := b.DeleteUpload(upload)
		if err != nil {
			b.log.Warningf("unable to delete expired upload : %s", err)
			errors = append(errors, err)
			continue
		}

		removed++
	}

	if len(errors) > 0 {
		return removed, fmt.Errorf("unable to delete %d expired uploads", len(errors))
	}

	return removed, nil
}