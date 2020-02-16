package metadata

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

// CreateToken create a new token in DB
func (b *Backend) CreateToken(token *common.Token) (err error) {
	return b.db.Create(token).Error
}

// GetToken return a token from the DB ( return nil and non error if not found )
func (b *Backend) GetToken(tokenStr string) (token *common.Token, err error) {
	token = &common.Token{}
	err = b.db.Where(&common.Token{Token: tokenStr}).Take(token).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return token, err
}

// GetTokens return all tokens for a user
func (b *Backend) GetTokens(user *common.User) (tokens []*common.Token, err error) {
	err = b.db.Where(&common.Token{UserID: user.ID}).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, err
}

// DeleteToken remove a token from the DB
func (b *Backend) DeleteToken(token *common.Token) error {

	// Delete token
	err := b.db.Delete(token).Error
	if err != nil {
		return fmt.Errorf("unable to delete token metadata")
	}

	return err
}
