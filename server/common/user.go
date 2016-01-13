package common
import "net/http"

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

func Logout(resp http.ResponseWriter){
	// Delete session cookie
	sessionCookie := &http.Cookie{}
	sessionCookie.HttpOnly = true
	//	secureCookie.Secure = true
	sessionCookie.Name = "plik-session"
	sessionCookie.Value = ""
	sessionCookie.MaxAge = -1
	sessionCookie.Path = "/"
	http.SetCookie(resp, sessionCookie)

	// Store xsrf token cookie
	xsrfCookie := &http.Cookie{}
	sessionCookie.HttpOnly = false
	//	secureCookie.Secure = true
	xsrfCookie.Name = "plik-xsrf"
	xsrfCookie.Value = ""
	xsrfCookie.MaxAge = -1
	xsrfCookie.Path = "/"
	http.SetCookie(resp, xsrfCookie)
}