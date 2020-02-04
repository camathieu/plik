package bolt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func TestBackend_CreateUser_MissingUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.CreateUser(nil)
	require.Errorf(t, err, "Missing user")
}

func TestBackend_CreateUser_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("users"))
	})
	require.NoError(t, err)

	user := common.NewUser()
	err = backend.CreateUser(user)
	require.Error(t, err)
}

func TestBackend_CreateUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	err := backend.CreateUser(user)
	require.NoError(t, err)
}

func TestBackend_CreateUser_Token(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"
	user.NewToken()

	err := backend.CreateUser(user)
	require.NoError(t, err)
}

func TestBackend_GetUser_NoUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	_, err := backend.GetUser("")
	require.Errorf(t, err, "Missing user")
}

func TestBackend_GetUser_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("users"))
	})
	require.NoError(t, err)

	_, err = backend.GetUser("id")
	require.Error(t, err)
}

func TestBackend_GetUser_NotFound(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user, err := backend.GetUser("id")
	require.NoError(t, err, "missing error")
	require.Nil(t, user, "invalid not nil user")
}

func TestBackend_GetUser_InvalidJSON(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	err := backend.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return errors.New("unable to get upload bucket")
		}

		err := bucket.Put([]byte(user.ID), []byte("invalid_json_value"))
		if err != nil {
			return errors.New("unable to put value")
		}

		return nil
	})
	require.NoError(t, err)

	_, err = backend.GetUser(user.ID)
	require.Error(t, err)
}

func TestBackend_GetUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	err := backend.CreateUser(user)
	require.NoError(t, err)

	_, err = backend.GetUser(user.ID)
	require.NoError(t, err)
}

func TestBackend_GetUser_ByToken(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"
	token := user.NewToken()

	err := backend.CreateUser(user)
	require.NoError(t, err, "save user error")

	u, err := backend.GetUserFromToken(token.Token)
	require.NoError(t, err, "unable to get user")
	require.NotNil(t, u, "invalid nil user")
	require.Equal(t, user.ID, u.ID, "invalid user")
}

func TestBackend_UpdateUser_NoUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user, err := backend.UpdateUser(nil, nil)
	common.RequireError(t, err, "missing user")
	require.Nil(t, user, "user is not nil")
}

func TestBackend_UpdateUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "plik"

	err := backend.CreateUser(user)
	require.NoError(t, err, "create user error")

	newID := "plok"
	tx := func(u *common.User) error {
		u.ID = newID
		return nil
	}

	u, err := backend.UpdateUser(user, tx)
	require.NoError(t, err, "missing user")
	require.NotNil(t, u, "user is nil")
	require.Equal(t, newID, u.ID, "user id mismatch")
}

func TestBackend_UpdateUser_TxError(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "plik"

	err := backend.CreateUser(user)
	require.NoError(t, err, "create user error")

	tx := func(u *common.User) error {
		if u == nil {
			return fmt.Errorf("no good")
		}
		return fmt.Errorf("tx error")
	}

	u, err := backend.UpdateUser(user, tx)
	common.RequireError(t, err, "tx error")
	require.Nil(t, u, "user is not nil")
}

func TestBackend_UpdateUser_NotFound(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "plik"

	tx := func(u *common.User) error {
		if u == nil {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("no good")
	}

	u, err := backend.UpdateUser(user, tx)
	common.RequireError(t, err, "user not found")
	require.Nil(t, u, "user is not nil")

	tx = func(u *common.User) error {
		return nil
	}

	u, err = backend.UpdateUser(user, tx)
	common.RequireError(t, err, "user tx without an user should return an error")
	require.Nil(t, u, "user is not nil")
}

func TestBackend_UpdateUser_InvalidJSON(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "plik"

	err := backend.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return errors.New("unable to get user bucket")
		}

		err := bucket.Put([]byte(user.ID), []byte("portnawak"))
		if err != nil {
			return errors.New("unable to put value")
		}

		return nil
	})
	require.NoError(t, err)

	tx := func(u *common.User) error {
		if u == nil {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("no good")
	}

	u, err := backend.UpdateUser(user, tx)
	common.RequireError(t, err, "unable to unserialize metadata from json")
	require.Nil(t, u, "user is not nil")
}

func TestBackend_RemoveUser_NoUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.RemoveUser(nil)
	require.Errorf(t, err, "Missing user")
}

func TestBackend_RemoveUser_NoBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("users"))
	})
	require.NoError(t, err, "unable to remove uploads bucket")

	user := common.NewUser()
	err = backend.RemoveUser(user)
	require.Error(t, err)
}

func TestBackend_RemoveUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"
	err := backend.CreateUser(user)
	require.NoError(t, err)

	err = backend.RemoveUser(user)
	require.NoError(t, err)

	u, err := backend.GetUser(user.ID)
	require.NoError(t, err)
	require.Nil(t, u, "non nil removed user")
}

func TestBackend_CreateUserToken_NoUser(t *testing.T) {

}

func TestBackend_CreateUserToken_NoToken(t *testing.T) {

}

func TestBackend_RemoveUser_NotFound(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	err := backend.RemoveUser(user)
	require.NoError(t, err)
}

func TestBackend_GetUserUploads_NoUser(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	_, err := backend.GetUserUploads(nil, nil)
	require.Errorf(t, err, "Missing user")
}

func TestBackend_GetUserUploads_NoBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("uploads"))
	})
	require.NoError(t, err)

	user := common.NewUser()
	_, err = backend.GetUserUploads(user, nil)
	require.Error(t, err)
}

func TestBackend_GetUserUploads(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	upload := common.NewUpload()
	upload.User = user.ID
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)

	upload2 := common.NewUpload()
	upload2.User = "another_user"
	upload2.Create()

	err = backend.CreateUpload(upload2)
	require.NoError(t, err)

	uploads, err := backend.GetUserUploads(user, nil)
	require.NoError(t, err, "get user error")
	require.Equal(t, 1, len(uploads), "invalid upload count")
	require.Equal(t, upload.ID, uploads[0], "invalid upload id")
}

func TestBackend_GetUserUploads_ByToken(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"
	token := user.NewToken()

	upload := common.NewUpload()
	upload.User = user.ID
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)

	upload2 := common.NewUpload()
	upload2.User = user.ID
	upload2.Token = token.Token
	upload2.Create()

	err = backend.CreateUpload(upload2)
	require.NoError(t, err)

	upload3 := common.NewUpload()
	upload3.User = "another_user"
	upload3.Create()

	err = backend.CreateUpload(upload3)
	require.NoError(t, err)

	uploads, err := backend.GetUserUploads(user, token)
	require.NoError(t, err)
	require.Equal(t, 1, len(uploads), "invalid upload count")
	require.Equal(t, upload2.ID, uploads[0], "invalid upload id")
}

func TestBackend_GetUsers_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("users"))
	})
	require.NoError(t, err)

	_, err = backend.GetUsers()
	require.Error(t, err)
}

func TestBackend_GetUsers(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user1 := common.NewUser()
	user1.ID = "ovh:test"
	user1.NewToken()

	err := backend.CreateUser(user1)
	require.NoError(t, err, "save user error")

	user2 := common.NewUser()
	user2.ID = "google:test"
	user2.NewToken()

	err = backend.CreateUser(user2)
	require.NoError(t, err, "save user error")

	users, err := backend.GetUsers()
	require.NoError(t, err, "get users error")
	require.Equal(t, 2, len(users), "invalid users count")
}

func TestBackend_GetUserStatistics(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"

	upload := common.NewUpload()
	upload.User = user.ID
	upload.Create()
	file1 := upload.NewFile()
	file1.CurrentSize = 1
	file2 := upload.NewFile()
	file2.CurrentSize = 2

	err := backend.CreateUpload(upload)
	require.NoError(t, err)

	upload2 := common.NewUpload()
	upload2.User = user.ID
	upload2.Create()
	file3 := upload2.NewFile()
	file3.CurrentSize = 3

	err = backend.CreateUpload(upload2)
	require.NoError(t, err)

	upload3 := common.NewUpload()
	upload3.Create()
	file4 := upload3.NewFile()
	file4.CurrentSize = 3000

	err = backend.CreateUpload(upload3)
	require.NoError(t, err)

	stats, err := backend.GetUserStatistics(user, nil)
	require.NoError(t, err, "get users error")
	require.Equal(t, 2, stats.Uploads, "invalid uploads count")
	require.Equal(t, 3, stats.Files, "invalid files count")
	require.Equal(t, int64(6), stats.TotalSize, "invalid file size")
}

func TestBackend_GetUserStatistics_ByToken(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	user := common.NewUser()
	user.ID = "user"
	token := user.NewToken()

	upload := common.NewUpload()
	upload.User = user.ID
	upload.Create()
	file1 := upload.NewFile()
	file1.CurrentSize = 1
	file2 := upload.NewFile()
	file2.CurrentSize = 2

	err := backend.CreateUpload(upload)
	require.NoError(t, err)

	upload2 := common.NewUpload()
	upload2.User = user.ID
	upload2.Token = token.Token
	upload2.Create()
	file3 := upload2.NewFile()
	file3.CurrentSize = 3

	err = backend.CreateUpload(upload2)
	require.NoError(t, err)

	upload3 := common.NewUpload()
	upload3.Create()
	file4 := upload3.NewFile()
	file4.CurrentSize = 3000

	err = backend.CreateUpload(upload3)
	require.NoError(t, err)

	stats, err := backend.GetUserStatistics(user, token)
	require.NoError(t, err, "get users error")
	require.Equal(t, 1, stats.Uploads, "invalid uploads count")
	require.Equal(t, 1, stats.Files, "invalid files count")
	require.Equal(t, int64(3), stats.TotalSize, "invalid file size")
}
