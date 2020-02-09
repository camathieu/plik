package metadata

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/root-gg/plik/server/common"
)

func (b *Backend) CreateToken(token *common.Token) (err error) {
	return b.db.Create(token).Error
}

func (b *Backend) GetToken(Token string) (token *common.Token, err error) {
	token = &common.Token{}
	err = b.db.Where(&common.Token{Token: Token}).Take(token).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return token, err
}

func (b *Backend) GetTokens(user *common.User) (tokens []*common.Token, err error) {
	err = b.db.Where(&common.Token{UserID: user.ID}).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, err
}

func (b *Backend) DeleteToken(token *common.Token) error {
	
	// Delete token
	err := b.db.Delete(token).Error
	if err != nil {
		return fmt.Errorf("unable to delete token metadata")
	}

	return err
}