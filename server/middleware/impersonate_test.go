package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func TestImpersonateNotAdmin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := httptest.NewRecorder()
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "You need administrator privileges")
}

func TestImpersonateMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetMetadataBackend(ctx).(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	user := common.NewUser()
	context.SetUser(ctx, user)
	context.SetAdmin(ctx, true)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := httptest.NewRecorder()
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to get user to impersonate")
}

func TestImpersonateUserNotFound(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	context.SetUser(ctx, user)
	context.SetAdmin(ctx, true)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := httptest.NewRecorder()
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusForbidden, "Unable to get user to impersonate : User does not exists")
}

func TestImpersonate(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	context.SetUser(ctx, user)
	context.SetAdmin(ctx, true)

	userToImpersonate := common.NewUser()
	userToImpersonate.ID = "user"
	err := context.GetMetadataBackend(ctx).CreateUser(userToImpersonate)
	require.NoError(t, err, "unable to save user to impersonate")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := httptest.NewRecorder()
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	userFromContext := context.GetUser(ctx)
	require.NotNil(t, userFromContext, "missing user from context")
	require.Equal(t, userToImpersonate, userFromContext, "invalid user from context")
}
