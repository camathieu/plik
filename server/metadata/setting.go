package metadata

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

// CreateSetting create a new setting in DB
func (b *Backend) CreateSetting(setting *common.Setting) (err error) {
	return b.db.Create(setting).Error
}

// GetSetting get a setting from DB
func (b *Backend) GetSetting(key string) (setting *common.Setting, err error) {
	setting = &common.Setting{}

	err = b.db.Take(setting, &common.Setting{Key: key}).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return setting, nil
}

// UpdateSetting update a setting in DB
func (b *Backend) UpdateSetting(key string, oldValue string, newValue string) (err error) {
	result := b.db.Where(&common.Setting{Key: key, Value: oldValue}).Update(&common.Setting{Value: newValue})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(1) {
		return fmt.Errorf("invalid file status")
	}

	return nil
}

// DeleteSetting delete a setting from DB
func (b *Backend) DeleteSetting(key string) (err error) {
	return b.db.Delete(&common.Setting{Key: key}).Error
}
