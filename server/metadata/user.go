package metadata

import (
	"fmt"
	"github.com/jinzhu/gorm"
	paginator "github.com/pilagod/gorm-cursor-paginator"
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

func (b *Backend) GetUsers(provider string, pagingQuery *common.PagingQuery) (users []*common.User, cursor *paginator.Cursor, err error) {
	if pagingQuery == nil {
		return nil, nil, fmt.Errorf("missing paging query")
	}

	p := pagingQuery.Paginator()

	p.SetKeys("Name", "CreatedAt")

	stmt := b.db.Model(&common.User{})
	err = p.Paginate(stmt, &users).Error
	if err != nil {
		return nil, nil, err
	}

	c := p.GetNextCursor()

	return users, &c, err
}

func (b *Backend) ForEachUserUpload(userID string, tokenStr string, f func(upload *common.Upload) error) (err error) {
	rows, err := b.db.Model(&common.Upload{}).Where(&common.Upload{User: userID, Token: tokenStr}).Rows()
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

func (b *Backend) DeleteUserUploads(userID string, tokenStr string) (removed int, err error) {

	deleted := 0
	var errors []error
	f := func(upload *common.Upload) (err error) {
		err = b.DeleteUpload(upload.ID)
		if err != nil {
			// TODO LOG
			errors = append(errors, err)
			return nil
		}
		deleted++
		return nil
	}

	err = b.ForEachUserUpload(userID, tokenStr, f)
	if err != nil {
		return deleted, err
	}
	if len(errors) > 0 {
		return deleted, fmt.Errorf("unable to delete all user uploads")
	}

	return deleted, nil
}

func (b *Backend) DeleteUser(userID string) (err error) {
	err = b.db.Transaction(func(tx *gorm.DB) (err error) {
		_, err = b.DeleteUserUploads(userID, "")
		if err != nil {
			return err
		}

		// Delete user tokens
		err = tx.Delete(&common.Token{UserID: userID}).Error
		if err != nil {
			return fmt.Errorf("unable to delete tokens metadata")
		}

		// Delete user
		err = tx.Unscoped().Delete(&common.User{ID: userID}).Error
		if err != nil {
			return fmt.Errorf("unable to delete user metadata")
		}

		return nil
	})

	return err
}
