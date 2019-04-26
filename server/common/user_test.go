/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015> Copyright holders list can be found in AUTHORS file
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	user := NewUser()
	require.NotNil(t, user, "invalid user")
	require.NotNil(t, user.Tokens, "invalid user")
}

func TestUserNewToken(t *testing.T) {
	user := NewUser()
	token := user.NewToken()
	require.NotNil(t, token, "missing token")
	require.NotZero(t, token.Token, "missing token initialization")
	require.NotZero(t, len(user.Tokens), "missing token")
	require.Equal(t, token, user.Tokens[0], "missing token")
}

func TestLogout(t *testing.T) {
	rr := httptest.NewRecorder()
	Logout(rr)
	require.Equal(t, 2, len(rr.Result().Cookies()), "missing response cookies")

	var sessionCookie *http.Cookie
	var xsrfCookie *http.Cookie

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "plik-session" {
			sessionCookie = cookie
		}
		if cookie.Name == "plik-xsrf" {
			xsrfCookie = cookie
		}
	}

	require.NotNil(t, sessionCookie, "missing session cookies")
	require.Equal(t, -1, sessionCookie.MaxAge, "invalid session cookies")

	require.NotNil(t, xsrfCookie, "missing xsrf cookies")
	require.Equal(t, -1, xsrfCookie.MaxAge, "invalid xsrf cookies")
}
