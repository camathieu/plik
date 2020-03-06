package metadata

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/root-gg/utils"
	"gopkg.in/gormigrate.v1"

	// load drivers
	// Still some issues to workaround for mysql
	//  - innodb deadlocks on create
	//  - createdAt => datetime(6)
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/root-gg/plik/server/common"
)

// Config metadata backend configuration
type Config struct {
	Driver           string
	ConnectionString string
	EraseFirst       bool
	Debug            bool
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

	b.db, err = gorm.Open(b.Config.Driver, b.Config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database : %s", err)
	}

	if config.Driver == "sqlite3" {
		err = b.db.Exec("PRAGMA journal_mode=WAL;").Error
		if err != nil {
			_ = b.db.Close()
			return nil, fmt.Errorf("unable to set wal mode : %s", err)
		}

		err = b.db.Exec("PRAGMA foreign_keys = ON").Error
		if err != nil {
			_ = b.db.Close()
			return nil, fmt.Errorf("unable to enable foreign keys : %s", err)
		}
	}

	if config.EraseFirst {
		err = b.db.DropTableIfExists("files", "uploads", "tokens", "users", "settings", "migrations").Error
		if err != nil {
			return nil, fmt.Errorf("unable to drop tables : %s", err)
		}
	}

	if config.Debug {
		b.db.LogMode(true)
	}

	err = b.initializeDB()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize DB : %s", err)
	}

	return b, err
}

func (b *Backend) initializeDB() (err error) {
	m := gormigrate.New(b.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// Create upload table
		{
			ID: "2020-03-21-12-00-00",
			Migrate: func(tx *gorm.DB) error {
				// Upload object
				type Upload struct {
					ID  string `json:"id"`
					TTL int    `json:"ttl"`

					DownloadDomain string `json:"downloadDomain"`
					RemoteIP       string `json:"uploadIp,omitempty"`
					Comments       string `json:"comments"`

					Files []*common.File `json:"files"`

					UploadToken string `json:"uploadToken,omitempty"`
					User        string `json:"user,omitempty" gorm:"index:idx_upload_user"`
					Token       string `json:"token,omitempty" gorm:"index:idx_upload_user_token"`
					IsAdmin     bool   `json:"admin"`

					Stream    bool `json:"stream"`
					OneShot   bool `json:"oneShot"`
					Removable bool `json:"removable"`

					ProtectedByPassword bool   `json:"protectedByPassword"`
					Login               string `json:"login,omitempty"`
					Password            string `json:"password,omitempty"`

					CreatedAt time.Time  `json:"createdAt"`
					DeletedAt *time.Time `json:"-" gorm:"index:idx_upload_deleted_at"`
					ExpireAt  *time.Time `json:"expireAt" gorm:"index:idx_upload_expire_at"`
				}
				return tx.AutoMigrate(&Upload{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("uploads").Error
			},
		},
		// Add upload column visibility
		{
			ID: "2020-03-21-12-38-00",
			Migrate: func(tx *gorm.DB) error {
				// Upload object
				type Upload struct {
					Visibility string `json:"visibility"`
				}
				return tx.AutoMigrate(&Upload{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Table("uploads").DropColumn("visibility").Error
			},
		},
	})

	m.InitSchema(func(tx *gorm.DB) error {

		//if b.Config.Driver == "mysql" {
		//	// Enable foreign keys
		//	tx = tx.Set("gorm:table_options", "ENGINE=InnoDB")
		//}

		err := tx.AutoMigrate(
			&common.Upload{},
			&common.File{},
			&common.User{},
			&common.Token{},
			&common.Setting{},
		).Error
		if err != nil {
			return err
		}

		//if b.Config.Driver == "mysql" {
		//	err = tx.Model(&common.File{}).AddForeignKey("upload_id", "uploads(id)", "RESTRICT", "RESTRICT").Error
		//	if err != nil {
		//		return err
		//	}
		//	err = tx.Model(&common.Token{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT").Error
		//	if err != nil {
		//		return err
		//	}
		//}

		// all other foreign keys...
		return nil
	})

	if err = m.Migrate(); err != nil {
		return fmt.Errorf("could not migrate: %v", err)
	}

	return nil
}

// Shutdown close the metadata backend
func (b *Backend) Shutdown() (err error) {
	return b.db.Close()
}
