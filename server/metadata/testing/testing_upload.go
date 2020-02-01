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

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

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

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

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

// UpdateUpload update upload metadata
func (b *Backend) UpdateUpload(upload *common.Upload, tx common.UploadTx) (u *common.Upload, err error) {
	if upload == nil {
		return nil, fmt.Errorf("missing upload")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	upload, ok := b.uploads[upload.ID]
	if !ok {
		err = tx(u)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("upload tx without upload should return an error")
	}

	u, err = defCopyUpload(upload)
	if err != nil {
		return nil, err
	}

	err = tx(u)
	if err != nil {
		return nil, err
	}

	u, err = defCopyUpload(u)
	if err != nil {
		return nil, err
	}

	// Avoid the possibility to override an other upload by changing the upload.ID in the tx
	b.uploads[upload.ID] = u

	u, err = defCopyUpload(u)
	if err != nil {
		return nil, err
	}

	return u, nil
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
