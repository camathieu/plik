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
		return nil, fmt.Errorf("unable to open Bolt database %s : %s", b.Config.Path, err)
	}

	// Create Bolt buckets if needed
	err = b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("uploads"))
		if err != nil {
			return fmt.Errorf("unable to create metadata bucket : %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("unable to create user bucket : %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Bolt buckets : %s", err)
	}

	return b, nil
}
