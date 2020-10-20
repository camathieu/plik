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
		err = b.db.DropTableIfExists("files", "uploads", "tokens", "users", "settings", "migrations", "invites").Error
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
		// you migrations here
		{
			ID: "initial",
			Migrate: func(tx *gorm.DB) error {
				// Initial database schema
				type File struct {
					ID       string `json:"id"`
					UploadID string `json:"-"  gorm:"type:varchar(255) REFERENCES uploads(id) ON UPDATE RESTRICT ON DELETE RESTRICT"`
					Name     string `json:"fileName"`

					Status string `json:"status"`

					Md5       string `json:"fileMd5"`
					Type      string `json:"fileType"`
					Size      int64  `json:"fileSize"`
					Reference string `json:"reference"`

					BackendDetails string `json:"-"`

					CreatedAt time.Time `json:"createdAt"`
				}

				type Upload struct {
					ID  string `json:"id"`
					TTL int    `json:"ttl"`

					DownloadDomain string `json:"downloadDomain"`
					RemoteIP       string `json:"uploadIp,omitempty"`
					Comments       string `json:"comments"`

					Files []*File `json:"files"`

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

				type Token struct {
					Token   string `json:"token" gorm:"primary_key"`
					Comment string `json:"comment,omitempty"`

					UserID string `json:"-" gorm:"type:varchar(255) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE CASCADE"`

					CreatedAt time.Time `json:"createdAt"`
				}

				type User struct {
					ID       string `json:"id,omitempty"`
					Provider string `json:"provider"`
					Login    string `json:"login,omitempty"`
					Password string `json:"-"`
					Name     string `json:"name,omitempty"`
					Email    string `json:"email,omitempty"`
					IsAdmin  bool   `json:"admin"`

					Tokens []*Token `json:"tokens,omitempty"`

					CreatedAt time.Time `json:"createdAt"`
				}

				type Setting struct {
					Key   string `gorm:"primary_key"`
					Value string
				}

				return tx.AutoMigrate(
					&Upload{},
					&File{},
					&User{},
					&Token{},
					&Setting{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("uploads", "files", "users", "tokens", "settings").Error
			},
		},
		{
			ID: "create-invite-table",
			Migrate: func(tx *gorm.DB) error {
				type Invite struct {
					ID     string  `json:"id,omitempty"`
					Issuer *string `json:"-" gorm:"type:varchar(255) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE CASCADE;index:idx_invite_issuer"`
					Admin  bool    `json:"admin"`

					ExpireAt  *time.Time `json:"expireAt" gorm:"index:idx_invite_expire_at"`
					CreatedAt time.Time  `json:"createdAt"`
				}
				return tx.AutoMigrate(&Invite{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("invites").Error
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
			&common.Invite{},
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
