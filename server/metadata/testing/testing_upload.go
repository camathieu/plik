package testing

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/root-gg/plik/server/common"
)

// CreateUpload create upload metadata
func (b *Backend) CreateUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return fmt.Errorf("missing upload")
	}

	if b.err != nil {
		return b.err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.uploads[upload.ID]; ok {
		return errors.New("upload already exists")
	}

	upload, err = defCopyUpload(upload)
	if err != nil {
		return err
	}

	b.uploads[upload.ID] = upload

	return nil
}

// GetUpload retrieve upload metadata
func (b *Backend) GetUpload(uploadID string) (upload *common.Upload, err error) {
	if uploadID == "" {
		return nil, fmt.Errorf("missing upload id")
	}

	if b.err != nil {
		return nil, b.err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	upload, ok := b.uploads[uploadID]
	if !ok {
		return nil, nil
	}

	upload, err = defCopyUpload(upload)
	if err != nil {
		return nil, err
	}

	return upload, nil
}

func (b *Backend) AddOrUpdateFile(upload *common.Upload, file *common.File, status string) (err error) {
	if upload == nil {
		return fmt.Errorf("missing upload")
	}

	if file == nil {
		return fmt.Errorf("missing upload id")
	}

	if b.err != nil {
		return b.err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	upload, ok := b.uploads[upload.ID]
	if !ok {
		return errors.New("upload does not exists anymore")
	}

	if status == "" {
		// Insert file, verify it does not already exist
		if _, ok := upload.Files[file.ID]; ok {
			return fmt.Errorf("file already exist")
		}
	} else {
		// Update file, verify it exists and status
		current, ok := upload.Files[file.ID]

		if !ok {
			return fmt.Errorf("missing file")
		}
		if current.Status != status {
			return fmt.Errorf("invalid file status %s, expected %s", current.Status, status)
		}

	}

	// Create a defensive copy
	f, err := defCopyFile(file)
	if err != nil {
		return err
	}

	// add file to upload
	upload.Files[file.ID] = f

	return nil
}

// RemoveUpload remove upload metadata
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return fmt.Errorf("missing upload")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	if _, ok := b.uploads[upload.ID]; !ok {
		return errors.New("upload does not exists")
	}

	delete(b.uploads, upload.ID)

	return nil
}

// Create a defensive copy of the upload object
func defCopyUpload(upload *common.Upload) (u *common.Upload, err error) {
	u = &common.Upload{}
	j, err := json.Marshal(upload)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, u)
	if err != nil {
		return nil, err
	}
	return u, err
}

// Create a defensive copy of the file object
func defCopyFile(file *common.File) (f *common.File, err error) {
	f = &common.File{}
	j, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, f)
	if err != nil {
		return nil, err
	}
	return f, err
}
