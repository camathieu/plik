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
	"time"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/utils"
)

// Ensure Bolt Metadata Backend implements metadata.Backend interface
var _ metadata.Backend = (*Backend)(nil)

// Config object
type Config struct {
	Path string
}

// NewConfig configures the backend from config passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	config.Path = "plik.db"
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	Config *Config

	db *bolt.DB
}

// NewBackend instantiate a new Bolt Metadata Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend, err error) {
	b = new(Backend)
	b.Config = config

	// Open the Bolt database
	b.db, err = bolt.Open(b.Config.Path, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("Unable to open Bolt database %s : %s", b.Config.Path, err)
	}

	// Create Bolt buckets if needed
	err = b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("uploads"))
		if err != nil {
			return fmt.Errorf("Unable to create metadata bucket : %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("Unable to create user bucket : %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to create Bolt buckets : %s", err)
	}

	return b, nil
}
