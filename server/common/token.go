package common

import (
	"fmt"
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
	t = &Token{}
	t.Initialize()
	return t
}

// Initialize generate the token uuid and sets the creation date
func (t *Token) Initialize() {
	t.CreationDate = time.Now().Unix()

	token, err := uuid.NewV4()
	if err != nil {
		panic(fmt.Errorf("unable to generate token uuid %s", err))
	}
	t.Token = token.String()
}
