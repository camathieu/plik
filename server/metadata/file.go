package metadata

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

// CreateFile persist a new file to the database
func (b *Backend) CreateFile(file *common.File) (err error) {
	return b.db.Create(file).Error
}

// GetFile return a file from the database ( nil and no error if not found )
func (b *Backend) GetFile(fileID string) (file *common.File, err error) {
	file = &common.File{}
	err = b.db.Where(&common.File{ID: fileID}).Take(file).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return file, err
}

// GetFiles return all files for the given upload ID
func (b *Backend) GetFiles(uploadID string) (files []*common.File, err error) {
	err = b.db.Where(&common.File{UploadID: uploadID}).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, err
}

// UpdateFile update a file in DB. Status ensure the file status has not changed since loaded
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

// UpdateFileStatus update a file status in DB. oldStatus ensure the file status has not changed since loaded
func (b *Backend) UpdateFileStatus(file *common.File, oldStatus string, newStatus string) error {
	result := b.db.Model(&common.File{}).Where(&common.File{ID: file.ID, Status: oldStatus}).Update(&common.File{Status: newStatus})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(1) {
		return fmt.Errorf("%s file not found", oldStatus)
	}

	file.Status = newStatus

	return nil
}

// RemoveFile change the file status to removed
// The file will then be deleted from the data backend by the server and the status changed to deleted.
func (b *Backend) RemoveFile(file *common.File) error {
	switch file.Status {
	case common.FileMissing, common.FileUploading, "":
		return b.UpdateFileStatus(file, file.Status, common.FileDeleted)
	case common.FileUploaded:
		return b.UpdateFileStatus(file, file.Status, common.FileRemoved)
	//case common.FileRemoved, common.FileDeleted:
	//	return nil
	default:
		return nil
	}
}

// ForEachUploadFiles execute f for each file of the upload
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

// ForEachRemovedFile execute f for each file with the status "removed"
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

// CountUploadFiles count how many files have been added to an upload
func (b *Backend) CountUploadFiles(uploadID string) (count int, err error) {
	err = b.db.Model(&common.File{}).Where(&common.File{UploadID: uploadID}).Count(&count).Error
	if err != nil {
		return -1, err
	}

	return count, nil
}
