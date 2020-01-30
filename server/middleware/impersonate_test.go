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

func TestImpersonateNotAdmin(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := ctx.NewRecorder(req)
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "you need administrator privileges")
}

func TestImpersonateMetadataBackendError(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	ctx.GetMetadataBackend().(*metadata_test.Backend).SetError(errors.New("metadata backend error"))

	user := common.NewUser()
	ctx.SetUser(user)
	ctx.SetAdmin(true)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := ctx.NewRecorder(req)

	context.TestPanic(t, rr, func() {
		Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)
	})
}

func TestImpersonateUserNotFound(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	ctx.SetUser(user)
	ctx.SetAdmin(true)

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := ctx.NewRecorder(req)
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestForbidden(t, rr, "user to impersonate does not exists")
}

func TestImpersonate(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	user := common.NewUser()
	ctx.SetUser(user)
	ctx.SetAdmin(true)

	userToImpersonate := common.NewUser()
	userToImpersonate.ID = "user"
	err := ctx.GetMetadataBackend().CreateUser(userToImpersonate)
	require.NoError(t, err, "unable to save user to impersonate")

	req, err := http.NewRequest("GET", "", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.Header.Set("X-Plik-Impersonate", "user")

	rr := ctx.NewRecorder(req)
	Impersonate(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	userFromContext := ctx.GetUser()
	require.NotNil(t, userFromContext, "missing user from context")
	require.Equal(t, userToImpersonate, userFromContext, "invalid user from context")
}
