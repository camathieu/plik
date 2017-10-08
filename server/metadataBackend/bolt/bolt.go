/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package bolt

import (
	"fmt"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/gob"

	"github.com/root-gg/plik/server/common"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
	"github.com/root-gg/utils"
)

// MetadataBackend object
type MetadataBackend struct {
	Config *MetadataBackendConfig

	db *storm.DB
}

// Until storms provides proper nested struct support
type Token struct {
	common.Token `storm:"inline"`
	UserId string `storm:"index"`
}

// NewBoltMetadataBackend instantiate a new Bolt Metadata Backend
// from configuration passed as argument
func NewBoltMetadataBackend(config map[string]interface{}) (bmb *MetadataBackend) {
	bmb = new(MetadataBackend)
	bmb.Config = NewBoltMetadataBackendConfig(config)

	// Open the Bolt database
	var err error
	bmb.db, err = storm.Open(bmb.Config.Path, storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}), storm.Codec(gob.Codec))
	if err != nil {
		log.Fatalf("Unable to open Bolt database %s : %s", bmb.Config.Path, err)
	}

	// Create Bolt buckets if needed
	err = bmb.db.Init(&common.User{})
	if err != nil {
		log.Fatalf("Unable to create user bucket : %s", err)
	}
	err = bmb.db.Init(&common.Upload{})
	if err != nil {
		log.Fatalf("Unable to create uploads bucket : %s", err)
	}
	err = bmb.db.Init(&Token{})
	if err != nil {
		log.Fatalf("Unable to create uploads bucket : %s", err)
	}

	return
}

// Create implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) SaveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return fmt.Errorf("Missing upload")
	}

	return bmb.db.Save(upload)
}

// Get implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) GetUpload(id string) (upload *common.Upload, err error) {
	if id == "" {
		err = fmt.Errorf("Missing upload id")
		return
	}

	upload = &common.Upload{}
	err = bmb.db.One("ID",id, upload)
	if err != nil {
		return nil, err
	}

	return upload, nil
}

// Remove implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return fmt.Errorf("Missing upload")
	}

	return bmb.db.DeleteStruct(upload)
}

// SaveUser implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) SaveUser(user *common.User) (err error) {
	if user == nil {
		return fmt.Errorf("Missing user")
	}

	common.Logger().Info("BEGIN TX")

	tx, err := bmb.db.Begin(true)
	if err != nil {
		return fmt.Errorf("Unable to create transaction : %s", err)
	}
	defer tx.Rollback()

	common.Logger().Info("SAVE USER")

	err = tx.Save(user)
	if err != nil {
		return fmt.Errorf("Unable to save user : %s", err)
	}

	common.Logger().Info("FIND")

	var tokens []Token
	err = tx.Find("UserId", user.ID, &tokens)
	if err != nil {
		if err != storm.ErrNotFound {
			return fmt.Errorf("Unable to get user tokens : %s", err)
		}
	}

	common.Logger().Info("ADD")

	// Add tokens
	for _, token := range user.Tokens {
		exists := false
		for _, t := range tokens {
			if token.ID == t.ID {
				exists = true
				break
			}
		}

		if !exists {
			// This token needs to be added
			common.Logger().Info("SAVING TOKEN")
			err = tx.Save(&Token{Token: *token, UserId: user.ID})
			if err != nil {
				return fmt.Errorf("Unable to add user token : %s", err)
			}
		}
	}

	common.Logger().Info("REMOVE")

	// Remove tokens
	for _, token := range tokens {
		exists := false
		for _, t := range user.Tokens {
			if token.ID == t.ID {
				exists = true
				break
			}
		}

		if !exists {
			// This token needs to be removed
			common.Logger().Info("REMOVING TOKEN")
			err = tx.Remove(&token)
			if err != nil {
				return fmt.Errorf("Unable to remove user token : %s", err)
			}
		}
	}

	common.Logger().Info("COMMIT")

	err = tx.Commit()
	if err != nil {
		return err
	}

	common.Logger().Info("DONE")

	return
}

// GetUser implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) GetUser(id string, token string) (user *common.User, err error) {
	if id == "" && token == "" {
		return nil, fmt.Errorf("Missing user id or token")
	}

	common.Logger().Info("token : " + token)

	if token != "" {
		t := &Token{}
		err = bmb.db.One("ID", token, t)
		if err != nil {
			return nil, fmt.Errorf("Unable to get token : %s", err)
		}
		common.Logger().Info(utils.Sdump(t))
		id = t.UserId
	}

	common.Logger().Info("user id : " + id)

	user = &common.User{}
	err = bmb.db.One("ID", id, user)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// RemoveUser implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) RemoveUser(user *common.User) (err error) {
	if user == nil {
		return fmt.Errorf("Missing user")
	}

	return bmb.db.DeleteStruct(user)
}

// GetUserUploads implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) GetUserUploads(user *common.User, token *common.Token) (ids []string, err error) {
	if user == nil {
		return nil, fmt.Errorf("Missing user")
	}

	matcher := q.Eq("User",user.ID)

	if token != nil {
		matcher = q.And(matcher, q.Eq("Token", token.ID))
	}

	var uploads []common.Upload
	err = bmb.db.Select(matcher).OrderBy("Creation").Reverse().Find(&uploads)
	if err != nil {
		if err != storm.ErrNotFound {
			return nil, err
		}
	}

	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}

	return ids, nil
}

// GetUploadsToRemove implementation for Bolt Metadata Backend
func (bmb *MetadataBackend) GetUploadsToRemove() (ids []string, err error) {
	var uploads []common.Upload
	err = bmb.db.Select(q.Lt("Deadline", time.Now().Unix())).Find(&uploads)
	if err != nil {
		if err != storm.ErrNotFound {
			return nil, err
		}
	}

	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}

	return ids, nil
}
