package common

import (
	"time"

	"github.com/nu7hatch/gouuid"
)

// Token provide a very basic authentication mechanism
type Token struct {
	Token        string `json:"token" bson:"token"`
	CreationDate int64  `json:"creationDate" bson:"creationDate"`
	Comment      string `json:"comment,omitempty" bson:"comment"`
}

// NewToken create a new Token instance
func NewToken() (t *Token) {
	t = new(Token)
	return
}

// Create initialize a new Token
func (t *Token) Create() (err error) {
	t.CreationDate = time.Now().Unix()
	uuid, err := uuid.NewV4()
	if err != nil {
		return
	}
	t.Token = uuid.String()
	return
}
