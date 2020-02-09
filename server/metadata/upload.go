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
	err = b.db.Take(upload, &common.Upload{ID: ID}).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return upload, err
}

func (b *Backend) RemoveUploadFiles(upload *common.Upload) (err error) {
	var errors []error
	f := func(file *common.File) (err error) {
		err = b.RemoveFile(file)
		if err != nil {
			errors = append(errors, err)
		}
		return nil
	}

	err = b.ForEachUploadFiles(upload, f)
	if err != nil {
		return err
	}
	if len(errors) > 0 {
		return fmt.Errorf("unable to remove %d files", len(errors))
	}

	return nil
}

func (b *Backend) DeleteUpload(upload *common.Upload) (err error) {
	err = b.db.Delete(upload).Error
	if err != nil {
		return fmt.Errorf("unable to delete upload")
	}

	err = b.RemoveUploadFiles(upload)
	if err != nil {
		return fmt.Errorf("unable to delete upload files")
	}

	return nil
}

func (b *Backend) DeleteExpiredUploads() (removed int, err error) {
	rows, err := b.db.Model(&common.Upload{}).Where("expire_at < ?", time.Now()).Rows()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch expired uploads : %s", err)
	}
	defer func() { _ = rows.Close() }()

	var errors []error
	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		if err != nil {
			return 0, fmt.Errorf("unable to fetch next expired upload : %s", err)
		}

		err := b.DeleteUpload(upload)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		removed++
	}

	if len(errors) > 0 {
		return removed, fmt.Errorf("unable to remove %d expired uploads", len(errors))
	}

	return removed, nil
}

func (b *Backend) PurgeDeletedUploads() (removed int, err error) {
	rows, err := b.db.Model(&common.Upload{}).Unscoped().Where("deleted_at IS NOT NULL").Rows()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch deletred uploads : %s", err)
	}
	defer func() { _ = rows.Close() }()

	var errors []error
	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		if err != nil {
			return removed, fmt.Errorf("unable to fetch next expired upload : %s", err)
		}

		var count int
		err := b.db.Model(&common.File{}).Not(&common.File{Status:common.FileDeleted}).Where(&common.File{UploadID: upload.ID}).Count(&count).Error
		if err != nil {
			return removed, err
		}
		if count > 0 {
			// TODO log properly
			fmt.Printf("Can't remove upload %s because %d files are still not deleted\n", upload.ID, count)
			continue
		}

		// Delete the upload files from the database
		err = b.db.Model(&common.File{}).Unscoped().Delete(&common.File{UploadID: upload.ID}).Error
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Delete the upload from the database
		err = b.db.Unscoped().Delete(upload).Error
		if err != nil {
			errors = append(errors, err)
			continue
		}
		removed++
	}

	if len(errors) > 0 {
		return removed, fmt.Errorf("unable to purge %d deleted uploads", len(errors))
	}

	return removed, nil
}