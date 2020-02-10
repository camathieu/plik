package metadata

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

func (b *Backend) CreateUser(user *common.User) (err error) {
	return b.db.Create(user).Error
}

func (b *Backend) GetUser(ID string) (user *common.User, err error) {
	user = &common.User{}
	err = b.db.Where(&common.User{ID: ID}).Take(user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return user, err
}

func (b *Backend) GetUserUploads(user *common.User, token *common.Token) (uploads []*common.Upload, err error) {
	whereClause := &common.Upload{User: user.ID}
	if token != nil {
		whereClause.Token = token.Token
	}

	err = b.db.Where(whereClause).Find(&uploads).Error
	if err != nil {
		return nil, err
	}
	return uploads, err
}

func (b *Backend) ForEachUserUpload(user *common.User, token *common.Token, f func(upload *common.Upload) error) (err error) {
	rows, err := b.db.Model(&common.Upload{}).Where(&common.Upload{User: user.ID}).Rows()
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		if err != nil {
			return err
		}
		err = f(upload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) DeleteUserUploads(user *common.User, token *common.Token) (removed int, err error) {

	deleted := 0
	var errors []error
	f := func(upload *common.Upload) (err error) {
		err = b.DeleteUpload(upload)
		if err != nil {
			// TODO LOG
			errors = append(errors, err)
			return nil
		}
		deleted++
		return nil
	}

	err = b.ForEachUserUpload(user, nil, f)
	if err != nil {
		return deleted, err
	}
	if len(errors) > 0 {
		return deleted, fmt.Errorf("unable to delete all user uploads")
	}

	return deleted, nil
}

func (b *Backend) DeleteUser(user *common.User) (err error) {
	err = b.db.Transaction(func(tx *gorm.DB) (err error) {
		_, err = b.DeleteUserUploads(user, nil)
		if err != nil {
			return err
		}

		// Delete user tokens
		err = tx.Delete(&common.Token{UserID: user.ID}).Error
		if err != nil {
			return fmt.Errorf("unable to delete tokens metadata")
		}

		// Delete user
		err = tx.Delete(user).Error
		if err != nil {
			return fmt.Errorf("unable to delete user metadata")
		}

		return nil
	})

	return err
}
