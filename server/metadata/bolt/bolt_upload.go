package bolt

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/common"
)

// CreateUpload implementation for Bolt Metadata Backend
func (b *Backend) CreateUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	// Serialize metadata to json
	j, err := json.Marshal(upload)
	if err != nil {
		return fmt.Errorf("unable to serialize metadata to json : %s", err)
	}

	// Save json metadata to Bolt database
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}

		if bucket.Get([]byte(upload.ID)) != nil {
			return fmt.Errorf("upload already exists")
		}

		err := bucket.Put([]byte(upload.ID), j)
		if err != nil {
			return fmt.Errorf("unable save upload metadata : %s", err)
		}

		// User index
		if upload.User != "" {
			// User index key is build as follow :
			//  - User index prefix 2 byte ( "_u" )
			//  - The user id
			//  - The upload date reversed ( 8 bytes )
			//  - The upload id ( 16 bytes )
			// Upload id is stored in the key to ensure uniqueness
			// AuthToken is stored in the value to permit byToken filtering
			timestamp := make([]byte, 8)
			binary.BigEndian.PutUint64(timestamp, ^uint64(0)-uint64(upload.Creation))

			key := append([]byte{'_', 'u'}, []byte(upload.User)...)
			key = append(key, timestamp...)
			key = append(key, []byte(upload.ID)...)

			err := bucket.Put(key, []byte(upload.Token))
			if err != nil {
				return fmt.Errorf("unable to save user index : %s", err)
			}
		}

		// Expire date index
		if upload.TTL > 0 {
			// Expire index is build as follow :
			//  - Expire index prefix 2 byte ( "_e" )
			//  - The expire timestamp ( 8 bytes )
			//  - The upload id ( 16 bytes )
			// Upload id is stored in the key to ensure uniqueness
			timestamp := make([]byte, 8)
			expiredTs := upload.Creation + int64(upload.TTL)
			binary.BigEndian.PutUint64(timestamp, uint64(expiredTs))

			key := append([]byte{'_', 'e'}, timestamp...)
			key = append(key, []byte(upload.ID)...)

			err := bucket.Put(key, []byte{})
			if err != nil {
				return fmt.Errorf("unable to save expire index : %s", err)
			}
		}

		return nil
	})
}

// GetUpload implementation for Bolt Metadata Backend
func (b *Backend) GetUpload(ID string) (upload *common.Upload, err error) {
	if ID == "" {
		return nil, errors.New("missing upload ID")
	}

	// Get json metadata from Bolt database
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}

		b := bucket.Get([]byte(ID))
		if b == nil || len(b) == 0 {
			return nil
		}

		// Unserialize metadata from json
		upload = new(common.Upload)
		err = json.Unmarshal(b, upload)
		if err != nil {
			return fmt.Errorf("unable to unserialize metadata from json \"%s\" : %s", string(b), err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return upload, nil
}

// UpdateUpload implementation for Bolt Metadata Backend
func (b *Backend) UpdateUpload(upload *common.Upload, uploadTx common.UploadTx) (u *common.Upload, err error) {
	if upload == nil {
		return nil, errors.New("missing upload")
	}

	// Remove upload from bolt database
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}

		b := bucket.Get([]byte(upload.ID))
		if b == nil || len(b) == 0 {
			// Upload not found ( maybe it has been removed in the mean time )
			// Let the upload tx set the (HTTP) error and forward it
			err = uploadTx(nil)
			if err != nil {
				return err
			}
			return fmt.Errorf("upload tx without an upload should return an error")
		}

		// Deserialize metadata from json
		u = new(common.Upload)
		err = json.Unmarshal(b, u)
		if err != nil {
			return fmt.Errorf("unable to unserialize metadata from json \"%s\" : %s", string(b), err)
		}

		// Mutate upload object
		err = uploadTx(u)
		if err != nil {
			return fmt.Errorf("unable to execute upload tx : %s", err)
		}

		// Serialize metadata to json
		j, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("unable to serialize metadata to json : %s", err)
		}

		// Avoid the possibility to override an other upload by changing the upload.ID in the tx
		err = bucket.Put([]byte(upload.ID), j)
		if err != nil {
			return fmt.Errorf("unable save upload metadata : %s", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return u, err
}

// RemoveUpload implementation for Bolt Metadata Backend
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	// Remove upload from bolt database
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}

		err := bucket.Delete([]byte(upload.ID))
		if err != nil {
			return err
		}

		// Remove upload user index
		if upload.User != "" {
			// User index key is build as follow :
			//  - User index prefix 2 byte ( "_u" )
			//  - The user id
			//  - The upload date reversed ( 8 bytes )
			//  - The upload id ( 16 bytes )
			// Upload id is stored in the key to ensure uniqueness
			// AuthToken is stored in the value to permit byToken filtering
			timestamp := make([]byte, 8)
			binary.BigEndian.PutUint64(timestamp, ^uint64(0)-uint64(upload.Creation))

			key := append([]byte{'_', 'u'}, []byte(upload.User)...)
			key = append(key, timestamp...)
			key = append(key, []byte(upload.ID)...)

			err := bucket.Delete(key)
			if err != nil {
				return fmt.Errorf("unable to delete user index : %s", err)
			}
		}

		// Remove upload expire date index
		if upload.TTL > 0 {
			// Expire index is build as follow :
			//  - Expire index prefix 2 byte ( "_e" )
			//  - The expire timestamp ( 8 bytes )
			//  - The upload id ( 16 bytes )
			// Upload id is stored in the key to ensure uniqueness
			timestamp := make([]byte, 8)
			expiredTs := upload.Creation + int64(upload.TTL)
			binary.BigEndian.PutUint64(timestamp, uint64(expiredTs))
			key := append([]byte{'_', 'e'}, timestamp...)
			key = append(key, []byte(upload.ID)...)

			err := bucket.Delete(key)
			if err != nil {
				return fmt.Errorf("unable to delete expire index : %s", err)
			}
		}

		return nil
	})
}
