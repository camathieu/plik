package bolt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/common"
)

// CreateUser implementation for Bolt Metadata Backend
func (b *Backend) CreateUser(user *common.User) (err error) {
	if user == nil {
		return errors.New("missing user")
	}

	// Serialize user to json
	j, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to serialize user to json : %s", err)
	}

	// Save json user to Bolt database
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		// Save user
		err := bucket.Put([]byte(user.ID), j)
		if err != nil {
			return fmt.Errorf("unable save user : %s", err)
		}

		// Update token index
		for _, token := range user.Tokens {
			err = bucket.Put([]byte(token.Token), []byte(user.ID))
			if err != nil {
				return fmt.Errorf("unable save new token index : %s", err)
			}
		}

		return nil
	})
}

// GetUser implementation for Bolt Metadata Backend
func (b *Backend) GetUser(userID string) (user *common.User, err error) {
	if userID == "" {
		return nil, errors.New("missing user ID")
	}

	// Get json user from Bolt database
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		b := bucket.Get([]byte(userID))

		// User not found but no error
		if b == nil || len(b) == 0 {
			return nil
		}

		// Unserialize user from json
		user = common.NewUser()
		err = json.Unmarshal(b, user)
		if err != nil {
			return fmt.Errorf("unable to unserialize user from json : %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get user : %s", err)
	}

	return user, nil
}

// GetUserFromToken implementation for Bolt Metadata Backend
func (b *Backend) GetUserFromToken(token string) (user *common.User, err error) {
	if token == "" {
		return nil, errors.New("missing token")
	}

	// Get json user from Bolt database
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		// token index lookup
		idBytes := bucket.Get([]byte(token))
		if idBytes == nil || len(idBytes) == 0 {
			return nil
		}
		userID := string(idBytes)

		b := bucket.Get([]byte(userID))

		// User not found but no error
		if b == nil || len(b) == 0 {
			return nil
		}

		// Unserialize user from json
		user = common.NewUser()
		err = json.Unmarshal(b, user)
		if err != nil {
			return fmt.Errorf("unable to unserialize user from json : %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get user : %s", err)
	}

	return user, nil
}

// UpdateUser implementation for Bolt Metadata Backend
func (b *Backend) UpdateUser(user *common.User, userTx common.UserTx) (u *common.User, err error) {
	if user == nil {
		return nil, errors.New("missing user")
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		b := bucket.Get([]byte(user.ID))

		// User not found but no error
		if b == nil || len(b) == 0 {
			// User not found ( maybe it has been removed in the mean time )
			// Let the upload tx set the (HTTP) error and forward it
			err = userTx(nil)
			if err != nil {
				return err
			}
			return fmt.Errorf("user tx without an user should return an error")
		}

		// Deserialize user from json
		u = common.NewUser()
		err = json.Unmarshal(b, u)
		if err != nil {
			return fmt.Errorf("unable to unserialize metadata from json : %s", err)
		}

		// Apply transaction ( mutate )
		err = userTx(u)
		if err != nil {
			return err
		}

		// Serialize user to json
		j, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("unable to serialize user to json : %s", err)
		}

		// Save user
		// Avoid the possibility to override an other user by changing the user.ID in the tx
		err = bucket.Put([]byte(user.ID), j)
		if err != nil {
			return fmt.Errorf("unable save user metadata : %s", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return u, err
}

// RemoveUser implementation for Bolt Metadata Backend
func (b *Backend) RemoveUser(user *common.User) (err error) {
	if user == nil {
		return errors.New("missing user")
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		err = bucket.Delete([]byte(user.ID))
		if err != nil {
			return err
		}

		// Update token index
		for _, token := range user.Tokens {
			err = bucket.Delete([]byte(token.Token))
			if err != nil {
				return fmt.Errorf("unable delete token index : %s", err)
			}
		}

		return nil
	})
}

// GetUserUploads implementation for Bolt Metadata Backend
func (b *Backend) GetUserUploads(user *common.User, token *common.Token) (ids []string, err error) {
	if user == nil {
		return nil, errors.New("missing user")

	}

	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return fmt.Errorf("unable to get uploads Bolt bucket")
		}
		cursor := bucket.Cursor()

		// User index key is build as follow :
		//  - User index prefix 2 byte ( "_u" )
		//  - The user id
		//  - The upload date reversed ( 8 bytes )
		//  - The upload id ( 16 bytes )
		// Upload id is stored in the key to ensure uniqueness
		// AuthToken is stored in the value to permit byToken filtering
		startKey := append([]byte{'_', 'u'}, []byte(user.ID)...)

		key, t := cursor.Seek(startKey)
		for key != nil && bytes.HasPrefix(key, startKey) {

			// byToken filter
			if token == nil || string(t) == token.Token {
				// Extract upload id from key ( 16 last bytes )
				ids = append(ids, string(key[len(key)-16:]))
			}

			// Scan the bucket forward
			key, t = cursor.Next()
		}

		return nil
	})

	return ids, err
}

// GetUsers implementation for Bolt Metadata Backend
func (b *Backend) GetUsers() (ids []string, err error) {
	// Get users from Bolt database
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("unable to get users Bolt bucket")
		}

		cursor := bucket.Cursor()

		for id, _ := cursor.First(); id != nil; id, _ = cursor.Next() {
			strid := string(id)

			// Discard tokens from the token index
			// TODO add an _ in front of the tokens
			if !(strings.HasPrefix(strid, "ovh") || strings.HasPrefix(strid, "google")) {
				continue
			}

			ids = append(ids, strid)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to get users : %s", err)
	}

	return ids, nil
}

// GetUserStatistics implementation for Bolt Metadata Backend
func (b *Backend) GetUserStatistics(user *common.User, token *common.Token) (stats *common.UserStats, err error) {
	stats = new(common.UserStats)

	ids, err := b.GetUserUploads(user, token)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		upload, err := b.GetUpload(id)
		if err != nil {
			continue
		}

		stats.Uploads++
		stats.Files += len(upload.Files)

		for _, file := range upload.Files {
			stats.TotalSize += file.CurrentSize
		}
	}

	return stats, nil
}
