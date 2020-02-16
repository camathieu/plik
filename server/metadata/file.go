package metadata

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

func (b *Backend) CreateFile(file *common.File) (err error) {
	return b.db.Create(file).Error
}

func (b *Backend) GetFile(ID string) (file *common.File, err error) {
	file = &common.File{}
	err = b.db.Where(&common.File{ID: ID}).Take(file).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return file, err
}

func (b *Backend) GetFiles(uploadID string) (files []*common.File, err error) {
	err = b.db.Where(&common.File{UploadID: uploadID}).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, err
}

func (b *Backend) UpdateFile(file *common.File, status string) error {
	result := b.db.Where(&common.File{ID: file.ID, Status: status}).Save(file)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(1) {
		return fmt.Errorf("invalid file status")
	}

	return nil
}

func (b *Backend) UpdateFileStatus(file *common.File, oldStatus string, newStatus string) error {
	result := b.db.Model(&common.File{}).Where(&common.File{ID: file.ID, Status: oldStatus}).Update(&common.File{Status: newStatus})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(1) {
		return fmt.Errorf("invalid file status")
	}

	file.Status = newStatus

	return nil
}

func (b *Backend) RemoveFile(file *common.File) error {
	switch file.Status {
	case common.FileRemoved, common.FileDeleted:
		return nil
	case common.FileMissing, common.FileUploading:
		return b.UpdateFileStatus(file, file.Status, common.FileDeleted)
	case common.FileUploaded:
		err := b.UpdateFileStatus(file, file.Status, common.FileRemoved)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) ForEachUploadFiles(uploadID string, f func(file *common.File) error) (err error) {
	rows, err := b.db.Model(&common.File{}).Where(&common.File{UploadID: uploadID}).Rows()
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		file := &common.File{}
		err = b.db.ScanRows(rows, file)
		if err != nil {
			return err
		}
		err = f(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) ForEachRemovedFile(f func(file *common.File) error) (err error) {
	rows, err := b.db.Model(&common.File{}).Where(&common.File{Status: common.FileRemoved}).Rows()
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		file := &common.File{}
		err = b.db.ScanRows(rows, file)
		if err != nil {
			return err
		}
		err = f(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) CountUploadFiles(uploadID string) (count int, err error) {
	err = b.db.Model(&common.File{}).Where(&common.File{UploadID: uploadID}).Count(&count).Error
	if err != nil {
		return -1, err
	}

	return count, nil
}
