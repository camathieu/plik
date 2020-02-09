package common

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/nu7hatch/gouuid"
)

// User is a plik user
type User struct {
	ID      string `json:"id,omitempty"`
	Login   string `json:"login,omitempty"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	IsAdmin bool   `json:"admin"`

	Tokens []*Token `json:"tokens,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt, omitempty"`
}

// NewUser create a new user object
func NewUser() *User {
	return &User{}
}

// NewToken add a new token to a user
func (user *User) NewToken() (token *Token) {
	token = NewToken()
	user.Tokens = append(user.Tokens, token)
	return token
}

// GenAuthCookies generate a sign a jwt session cookie to authenticate a user
func GenAuthCookies(user *User, config *Configuration) (sessionCookie *http.Cookie, xsrfCookie *http.Cookie, err error) {
	var provider string
	var sig string
	if strings.HasPrefix(user.ID, "ovh:") {
		provider = "ovh"
		sig = config.OvhAPISecret
	} else if strings.HasPrefix(user.ID, "google:") {
		provider = "google"
		sig = config.GoogleAPISecret
	} else {
		return nil, nil, fmt.Errorf("invlid user id from unknown provider")
	}

	// Generate session jwt
	session := jwt.New(jwt.SigningMethodHS256)
	session.Claims.(jwt.MapClaims)["uid"] = user.ID
	session.Claims.(jwt.MapClaims)["provider"] = provider

	// Generate xsrf token
	xsrfToken, err := uuid.NewV4()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate xsrf token")
	}
	session.Claims.(jwt.MapClaims)["xsrf"] = xsrfToken.String()

	sessionString, err := session.SignedString([]byte(sig))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign session cookie : %s", err)
	}

	// Store session jwt in secure cookie
	sessionCookie = &http.Cookie{}
	sessionCookie.HttpOnly = true
	sessionCookie.Secure = true
	sessionCookie.Name = "plik-session"
	sessionCookie.Value = sessionString
	sessionCookie.MaxAge = int(time.Now().Add(10 * 365 * 24 * time.Hour).Unix())
	sessionCookie.Path = "/"

	// Store xsrf token cookie
	xsrfCookie = &http.Cookie{}
	xsrfCookie.HttpOnly = false
	xsrfCookie.Secure = true
	xsrfCookie.Name = "plik-xsrf"
	xsrfCookie.Value = xsrfToken.String()
	xsrfCookie.MaxAge = int(time.Now().Add(10 * 365 * 24 * time.Hour).Unix())
	xsrfCookie.Path = "/"

	return sessionCookie, xsrfCookie, nil
}

// ParseSessionCookie parse and validate the session cookie
func ParseSessionCookie(value string, config *Configuration) (uid string, xsrf string, err error) {
	session, err := jwt.Parse(value, func(t *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected siging method : %v", t.Header["alg"])
		}

		// Get authentication provider
		provider, ok := t.Claims.(jwt.MapClaims)["provider"]
		if !ok {
			return nil, fmt.Errorf("missing authentication provider")
		}

		switch provider {
		case "google":
			if config.GoogleAPISecret == "" {
				return nil, fmt.Errorf("missing Google API credentials")
			}
			return []byte(config.GoogleAPISecret), nil
		case "ovh":
			if config.OvhAPISecret == "" {
				return nil, fmt.Errorf("missing OVH API credentials")
			}
			return []byte(config.OvhAPISecret), nil
		default:
			return nil, fmt.Errorf("invalid authentication provider : %s", provider)
		}
	})
	if err != nil {
		return "", "", err
	}

	// Get the user id
	userValue, ok := session.Claims.(jwt.MapClaims)["uid"]
	if ok {
		uid, ok = userValue.(string)
		if !ok || uid == "" {
			return "", "", fmt.Errorf("missing user from session cookie")
		}
	} else {
		return "", "", fmt.Errorf("missing user from session cookie")
	}

	// Get the xsrf token
	xsrfValue, ok := session.Claims.(jwt.MapClaims)["xsrf"]
	if ok {
		xsrf, ok = xsrfValue.(string)
		if !ok || uid == "" {
			return "", "", fmt.Errorf("missing xsrf token from session cookie")
		}
	} else {
		return "", "", fmt.Errorf("missing xsrf token from session cookie")
	}

	return uid, xsrf, nil
}

// Logout delete plik session cookies
func Logout(resp http.ResponseWriter) {
	// Delete session cookie
	sessionCookie := &http.Cookie{}
	sessionCookie.HttpOnly = true
	sessionCookie.Secure = true
	sessionCookie.Name = "plik-session"
	sessionCookie.Value = ""
	sessionCookie.MaxAge = -1
	sessionCookie.Path = "/"
	http.SetCookie(resp, sessionCookie)

	// Store xsrf token cookie
	xsrfCookie := &http.Cookie{}
	xsrfCookie.HttpOnly = false
	xsrfCookie.Secure = true
	xsrfCookie.Name = "plik-xsrf"
	xsrfCookie.Value = ""
	xsrfCookie.MaxAge = -1
	xsrfCookie.Path = "/"
	http.SetCookie(resp, xsrfCookie)
}
