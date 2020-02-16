package common

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/nu7hatch/gouuid"
)

// ProviderGoogle for authentication
const ProviderGoogle = "google"

// ProviderOVH for authentication
const ProviderOVH = "ovh"

// ProviderLocal for authentication
const ProviderLocal = "local"

// User is a plik user
type User struct {
	ID       string `json:"id,omitempty"`
	Provider string `json:"provider"`
	Login    string `json:"login,omitempty"`
	Password string `json:"-"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	IsAdmin  bool   `json:"admin"`

	Tokens []*Token `json:"tokens,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// NewUser create a new user object
func NewUser(provider string, providerID string) (user *User) {
	user = &User{}
	user.ID = GetUserId(provider, providerID)
	user.Provider = provider
	return user
}

func GetUserId(provider string, providerID string) string {
	return fmt.Sprintf("%s:%s", provider, providerID)
}

// NewToken add a new token to a user
func (user *User) NewToken() (token *Token) {
	token = NewToken()
	token.UserID = user.ID
	user.Tokens = append(user.Tokens, token)
	return token
}

// NewToken add a new token to a user
func (user *User) String() string {
	str := user.Provider + ":" + user.Login
	if user.Name != "" {
		str += " " + user.Name
	}
	if user.Email != "" {
		str += " " + user.Email
	}
	return str
}

func (config *Configuration) getSignatureKey(provider string) (key string, err error) {
	switch provider {
	case ProviderGoogle:
		if config.GoogleAPISecret == "" {
			return "", fmt.Errorf("Google authentication is disabled")
		}
		return config.GoogleAPISecret, nil
	case ProviderOVH:
		if config.OvhAPISecret == "" {
			return "", fmt.Errorf("OVH authentication is disabled")
		}
		return config.OvhAPISecret, nil
	case ProviderLocal:
		// TODO
		return "TODO", nil
	default:
		return "", fmt.Errorf("unknown authentication provider")
	}
}

// GenAuthCookies generate a sign a jwt session cookie to authenticate a user
func GenAuthCookies(user *User, config *Configuration) (sessionCookie *http.Cookie, xsrfCookie *http.Cookie, err error) {
	// Generate session jwt
	session := jwt.New(jwt.SigningMethodHS256)
	session.Claims.(jwt.MapClaims)["uid"] = user.ID
	session.Claims.(jwt.MapClaims)["provider"] = user.Provider

	// Generate xsrf token
	xsrfToken, err := uuid.NewV4()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate xsrf token")
	}
	session.Claims.(jwt.MapClaims)["xsrf"] = xsrfToken.String()

	signatureKey, err := config.getSignatureKey(user.Provider)
	if err != nil {
		return nil, nil, err
	}

	sessionString, err := session.SignedString([]byte(signatureKey))
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
		providerValue, ok := t.Claims.(jwt.MapClaims)["provider"]
		if !ok {
			return nil, fmt.Errorf("missing authentication provider")
		}

		provider, ok := providerValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid authentication provider")
		}

		signatureKey, err := config.getSignatureKey(provider)
		if err != nil {
			return nil, err
		}

		return []byte(signatureKey), nil
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}