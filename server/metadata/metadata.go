package metadata

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/utils"
	"gopkg.in/gormigrate.v1"
)

// Config describes configuration for File Databackend
type Config struct {
	Driver           string
	ConnectionString string
	EraseFirst       bool
}

// NewConfig instantiate a new default configuration
// and override it with configuration passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	config.Driver = "sqlite3"
	config.ConnectionString = "plik.db"
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	Config *Config

	db *gorm.DB
}

// NewBackend instantiate a new File Data Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend, err error) {
	b = new(Backend)
	b.Config = config

	if config.Driver == "sqlite3" && config.EraseFirst {
		_ = os.Remove(config.ConnectionString)
	}

	b.db, err = gorm.Open(b.Config.Driver, b.Config.ConnectionString)
	if err != nil {
		return nil, err
	}

	if config.Driver == "sqlite3" {
		err = b.db.Exec("PRAGMA journal_mode=WAL;").Error
		if err != nil {
			_ = b.db.Close()
			return nil, err
		}

		err = b.db.Exec("PRAGMA foreign_keys = ON").Error
		if err != nil {
			_ = b.db.Close()
			return nil, err
		}
	}

	err = b.initializeDB()
	if err != nil {
		return nil, err
	}

	return b, err
}

func (b *Backend) initializeDB() (err error){
	m := gormigrate.New(b.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// you migrations here
	})

	m.InitSchema(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(
			&common.Upload{},
			&common.File{},
			&common.User{},
			&common.Token{},
		).Error
		if err != nil {
			return err
		}

		if b.Config.Driver == "mysql" {
			err = tx.Model(&common.Upload{}).AddForeignKey("user", "users(id)", "RESTRICT", "RESTRICT").Error
			if err != nil {
				return err
			}
			err = tx.Model(&common.Upload{}).AddForeignKey("token", "tokens(token)", "RESTRICT", "RESTRICT").Error
			if err != nil {
				return err
			}
		}

		// all other foreign keys...
		return nil

	})

	if err = m.Migrate(); err != nil {
		return fmt.Errorf("could not migrate: %v", err)
	}

	return nil
}

func (b *Backend) Shutdown() (err error) {
	return b.db.Close()
}
