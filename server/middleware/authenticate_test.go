package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateTokenNoUser(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-PlikToken", "token")

	rr := ctx.NewRecorder(req)
	Authenticate(true)(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "invalid token")
}

func TestAuthenticateTokenMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true
	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-PlikToken", "token")

	rr := ctx.NewRecorder(req)

	f := func() {
		Authenticate(true)(ctx, common.DummyHandler).ServeHTTP(rr, req)
	}
	context.TestPanic(t, rr, f)
}

func TestAuthenticateToken(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	user := common.NewUser()
	token := user.NewToken()

	err := ctx.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to save user to impersonate")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-PlikToken", token.Token)

	rr := ctx.NewRecorder(req)
	Authenticate(true)(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	userFromContext := ctx.GetUser()
	tokenFromContext := ctx.GetToken()
	require.Equal(t, user, userFromContext, "missing user from context")
	require.Equal(t, token, tokenFromContext, "invalid token from context")
}

func TestAuthenticateInvalidSessionCookie(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	cookie := &http.Cookie{}
	cookie.Name = "plik-session"
	cookie.Value = "invalid_value"
	req.AddCookie(cookie)

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "invalid session")
}

func TestAuthenticateMissingXSRFHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true
	ctx.GetConfig().OvhAuthentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"

	req, err := http.NewRequest("POST", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "missing xsrf header")
}

func TestAuthenticateInvalidXSRFHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().Authentication = true
	ctx.GetConfig().OvhAuthentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"

	req, err := http.NewRequest("POST", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	req.Header.Set("X-XSRFToken", "invalid_header_value")

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "invalid xsrf header")
}

func TestAuthenticateMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().OvhAuthentication = true
	ctx.GetConfig().Authentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"
	err := ctx.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to save user")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	rr := ctx.NewRecorder(req)
	context.TestPanic(t, rr, func() {
		Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)
	})
}

func TestAuthenticateNoUser(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().OvhAuthentication = true
	ctx.GetConfig().Authentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)
	context.TestForbidden(t, rr, "invalid session : user does not exists")

}

func TestAuthenticate(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().OvhAuthentication = true
	ctx.GetConfig().Authentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"
	err := ctx.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to save user")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, user, ctx.GetUser(), "invalid user from context")
}

func TestAuthenticateAdminUser(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetConfig().OvhAuthentication = true
	ctx.GetConfig().Authentication = true

	key := "secret_key"
	ctx.GetConfig().OvhAPISecret = key

	user := common.NewUser()
	user.ID = "ovh:user"
	err := ctx.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to save user")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	// Generate session cookie
	sessionCookie, _, err := common.GenAuthCookies(user, ctx.GetConfig())
	require.NoError(t, err, "unable to create new request")
	req.AddCookie(sessionCookie)

	rr := ctx.NewRecorder(req)
	Authenticate(false)(ctx, common.DummyHandler).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, user, ctx.GetUser(), "invalid user from context")
}
