package metadata

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func createUser(t *testing.T, b *Backend, user *common.User) {
    err := b.CreateUser(user)
    require.NoError(t, err, "create user error : %s", err)
}

func TestBackend_CreateUser(t *testing.T) {
	b := newTestMetadataBackend()

	user := &common.User{ID: "user"}
    createUser(t, b, user)
	require.NotZero(t, user.ID, "missing user id")
	require.NotZero(t, user.CreatedAt, "missing creation date")
}


func TestBackend_GetUser(t *testing.T) {
	b := newTestMetadataBackend()

	user := &common.User{ID: "user"}
	createUser(t, b, user)

	result, err := b.GetUser(user.ID)
	require.NoError(t, err, "get user error")
	require.Equal(t, user.ID, result.ID, "invalid user id")
}


func TestBackend_GetUser_NotFound(t *testing.T) {
	b := newTestMetadataBackend()

	user, err := b.GetUser("not found")
	require.NoError(t, err, "get user error")
	require.Nil(t, user, "user not nil")
}


func TestBackend_DeleteUser(t *testing.T) {
	b := newTestMetadataBackend()

	user := &common.User{}

    createUser(t, b, user)

	err := b.DeleteUser(user)
	require.NoError(t, err, "get user error")


	user, err = b.GetUser(user.ID)
	require.NoError(t, err, "get user error")
	require.Nil(t, user, "user not nil")
}