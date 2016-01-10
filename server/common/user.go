package common

type User struct {
	ID    string `json:"id,omitempty" bson:"id"`
	Name  string `json:"name,omitempty" bson:"name"`
	Email string `json:"email,omitempty" bson:"email"`
	//	Password string	`json:"password,omitempty" bson:"password"` // NYI
	Tokens []*Token `json:"tokens,omitempty" bson:"tokens"`
	//	IsAdmin bool 	`json:"isAdmin,omitempty" bson:"isAdmnin"`// NYI
}

func NewUser() (user *User) {
	user = new(User)
	user.Tokens = make([]*Token, 0)
	return
}

func (user *User) NewToken() (token *Token) {
	token = NewToken()
	token.Create()
	user.Tokens = append(user.Tokens, token)
	return
}
