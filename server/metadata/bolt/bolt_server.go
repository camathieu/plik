package bolt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/common"
)

// GetUploadsToRemove implementation for Bolt Metadata Backend
func (b *Backend) GetUploadsToRemove() (ids []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}
		cursor := bucket.Cursor()

		// Expire index is build as follow :
		//  - Expire index prefix 2 byte ( "_e" )
		//  - The expire timestamp ( 8 bytes )
		//  - The upload id ( 16 bytes )
		// Upload id is stored in the key to ensure uniqueness

		// Create seek key at current timestamp + 1
		timestamp := make([]byte, 8)
		binary.BigEndian.PutUint64(timestamp, uint64(time.Now().Unix()+1))
		startKey := append([]byte{'_', 'e'}, timestamp...)

		// Seek just after the seek key
		// All uploads above the cursor are expired
		cursor.Seek(startKey)
		for {
			// Scan the bucket upwards
			key, _ := cursor.Prev()
			if key == nil || !bytes.HasPrefix(key, []byte("_e")) {
				break
			}

			// Extract upload id from key ( 16 last bytes )
			ids = append(ids, string(key[10:]))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// GetServerStatistics implementation for Bolt Metadata Backend
func (b *Backend) GetServerStatistics() (stats *common.ServerStats, err error) {
	stats = new(common.ServerStats)

	// Get ALL upload ids
	var ids []string
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}
		cursor := bucket.Cursor()

		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			// Ignore indexes
			if bytes.HasPrefix(key, []byte("_")) {
				continue
			}

			ids = append(ids, string(key))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Compute upload statistics

	byTypeAggregator := common.NewByTypeAggregator()

	for _, id := range ids {
		upload, err := b.GetUpload(id)
		if upload == nil || err != nil {
			continue
		}

		stats.AddUpload(upload)

		for _, file := range upload.Files {
			byTypeAggregator.AddFile(file)
		}
	}

	stats.FileTypeByCount = byTypeAggregator.GetFileTypeByCount(10)
	stats.FileTypeBySize = byTypeAggregator.GetFileTypeBySize(10)

	// User count
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}
		cursor := bucket.Cursor()

		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			stats.Users++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return stats, nil
}
