package bolt

//
//import (
//	"errors"
//	"testing"
//
//	"github.com/boltdb/bolt"
//	"github.com/root-gg/plik/server/common"
//	"github.com/root-gg/plik/server/context"
//	"github.com/stretchr/testify/require"
//)
//
//func TestBackend_CreateUser_MissingUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.CreateUser(ctx, nil)
//	require.Errorf(t, err, "Missing user")
//}
//
//func TestBackend_CreateUser_MissingBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("users"))
//	})
//	require.NoError(t, err)
//
//	user := common.NewUser()
//	err = backend.CreateUser(ctx, user)
//	require.Error(t, err)
//}
//
//func TestBackend_CreateUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	err := backend.CreateUser(ctx, user)
//	require.NoError(t, err)
//}
//
//func TestBackend_CreateUser_Token(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//	user.NewToken()
//
//	err := backend.CreateUser(ctx, user)
//	require.NoError(t, err)
//}
//
//func TestBackend_GetUser_NoUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	_, err := backend.GetUser(ctx, "", "")
//	require.Errorf(t, err, "Missing user")
//}
//
//func TestBackend_GetUser_MissingBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("users"))
//	})
//	require.NoError(t, err)
//
//	_, err = backend.GetUser(ctx, "id", "")
//	require.Error(t, err)
//}
//
//func TestBackend_GetUser_NotFound(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user, err := backend.GetUser(ctx, "id", "")
//	require.NoError(t, err, "missing error")
//	require.Nil(t, user, "invalid not nil user")
//}
//
//func TestBackend_GetUser_InvalidJSON(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("users"))
//		if bucket == nil {
//			return errors.New("unable to get upload bucket")
//		}
//
//		err := bucket.Put([]byte(user.ID), []byte("invalid_json_value"))
//		if err != nil {
//			return errors.New("unable to put value")
//		}
//
//		return nil
//	})
//	require.NoError(t, err)
//
//	_, err = backend.GetUser(ctx, user.ID, "")
//	require.Error(t, err)
//}
//
//func TestBackend_GetUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	err := backend.CreateUser(ctx, user)
//	require.NoError(t, err)
//
//	_, err = backend.GetUser(ctx, user.ID, "")
//	require.NoError(t, err)
//}
//
//func TestBackend_GetUser_ByToken(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//	token := user.NewToken()
//
//	err := backend.CreateUser(ctx, user)
//	require.NoError(t, err, "save user error")
//
//	u, err := backend.GetUser(ctx, "", token.Token)
//	require.NoError(t, err, "unable to get user")
//	require.NotNil(t, u, "invalid nil user")
//	require.Equal(t, user.ID, u.ID, "invalid user")
//}
//
//func TestBackend_RemoveUser_NoUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.RemoveUser(ctx, nil)
//	require.Errorf(t, err, "Missing user")
//}
//
//func TestBackend_RemoveUser_NoBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("users"))
//	})
//	require.NoError(t, err, "unable to remove uploads bucket")
//
//	user := common.NewUser()
//	err = backend.RemoveUser(ctx, user)
//	require.Error(t, err)
//}
//
//func TestBackend_RemoveUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//	err := backend.CreateUser(ctx, user)
//	require.NoError(t, err)
//
//	err = backend.RemoveUser(ctx, user)
//	require.NoError(t, err)
//
//	u, err := backend.GetUser(ctx, user.ID, "")
//	require.NoError(t, err)
//	require.Nil(t, u, "non nil removed user")
//}
//
//func TestBackend_CreateUserToken_NoUser(t *testing.T) {
//
//}
//
//func TestBackend_CreateUserToken_NoToken(t *testing.T) {
//
//}
//
//func TestBackend_RemoveUser_NotFound(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	err := backend.RemoveUser(ctx, user)
//	require.NoError(t, err)
//}
//
//func TestBackend_GetUserUploads_NoUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	_, err := backend.GetUserUploads(ctx, nil, nil)
//	require.Errorf(t, err, "Missing user")
//}
//
//func TestBackend_GetUserUploads_NoBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("uploads"))
//	})
//	require.NoError(t, err)
//
//	user := common.NewUser()
//	_, err = backend.GetUserUploads(ctx, user, nil)
//	require.Error(t, err)
//}
//
//func TestBackend_GetUserUploads(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	upload := common.NewUpload()
//	upload.User = user.ID
//	upload.Create()
//
//	err := backend.CreateUpload(ctx, upload)
//	require.NoError(t, err)
//
//	upload2 := common.NewUpload()
//	upload2.User = "another_user"
//	upload2.Create()
//
//	err = backend.CreateUpload(ctx, upload2)
//	require.NoError(t, err)
//
//	uploads, err := backend.GetUserUploads(ctx, user, nil)
//	require.NoError(t, err, "get user error")
//	require.Equal(t, 1, len(uploads), "invalid upload count")
//	require.Equal(t, upload.ID, uploads[0], "invalid upload id")
//}
//
//func TestBackend_GetUserUploads_ByToken(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//	token := user.NewToken()
//
//	upload := common.NewUpload()
//	upload.User = user.ID
//	upload.Create()
//
//	err := backend.CreateUpload(ctx, upload)
//	require.NoError(t, err)
//
//	upload2 := common.NewUpload()
//	upload2.User = user.ID
//	upload2.Token = token.Token
//	upload2.Create()
//
//	err = backend.CreateUpload(ctx, upload2)
//	require.NoError(t, err)
//
//	upload3 := common.NewUpload()
//	upload3.User = "another_user"
//	upload3.Create()
//
//	err = backend.CreateUpload(ctx, upload3)
//	require.NoError(t, err)
//
//	uploads, err := backend.GetUserUploads(ctx, user, token)
//	require.NoError(t, err)
//	require.Equal(t, 1, len(uploads), "invalid upload count")
//	require.Equal(t, upload2.ID, uploads[0], "invalid upload id")
//}
//
//func TestBackend_GetUsers_MissingBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("users"))
//	})
//	require.NoError(t, err)
//
//	_, err = backend.GetUsers(ctx)
//	require.Error(t, err)
//}
//
//func TestBackend_GetUsers(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user1 := common.NewUser()
//	user1.ID = "ovh:test"
//	user1.NewToken()
//
//	err := backend.CreateUser(ctx, user1)
//	require.NoError(t, err, "save user error")
//
//	user2 := common.NewUser()
//	user2.ID = "google:test"
//	user2.NewToken()
//
//	err = backend.CreateUser(ctx, user2)
//	require.NoError(t, err, "save user error")
//
//	users, err := backend.GetUsers(ctx)
//	require.NoError(t, err, "get users error")
//	require.Equal(t, 2, len(users), "invalid users count")
//}
//
//func TestBackend_GetUserStatistics(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//
//	upload := common.NewUpload()
//	upload.User = user.ID
//	upload.Create()
//	file1 := upload.NewFile()
//	file1.CurrentSize = 1
//	file2 := upload.NewFile()
//	file2.CurrentSize = 2
//
//	err := backend.CreateUpload(ctx, upload)
//	require.NoError(t, err)
//
//	upload2 := common.NewUpload()
//	upload2.User = user.ID
//	upload2.Create()
//	file3 := upload2.NewFile()
//	file3.CurrentSize = 3
//
//	err = backend.CreateUpload(ctx, upload2)
//	require.NoError(t, err)
//
//	upload3 := common.NewUpload()
//	upload3.Create()
//	file4 := upload3.NewFile()
//	file4.CurrentSize = 3000
//
//	err = backend.CreateUpload(ctx, upload3)
//	require.NoError(t, err)
//
//	stats, err := backend.GetUserStatistics(ctx, user, nil)
//	require.NoError(t, err, "get users error")
//	require.Equal(t, 2, stats.Uploads, "invalid uploads count")
//	require.Equal(t, 3, stats.Files, "invalid files count")
//	require.Equal(t, int64(6), stats.TotalSize, "invalid file size")
//}
//
//func TestBackend_GetUserStatistics_ByToken(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	user := common.NewUser()
//	user.ID = "user"
//	token := user.NewToken()
//
//	upload := common.NewUpload()
//	upload.User = user.ID
//	upload.Create()
//	file1 := upload.NewFile()
//	file1.CurrentSize = 1
//	file2 := upload.NewFile()
//	file2.CurrentSize = 2
//
//	err := backend.CreateUpload(ctx, upload)
//	require.NoError(t, err)
//
//	upload2 := common.NewUpload()
//	upload2.User = user.ID
//	upload2.Token = token.Token
//	upload2.Create()
//	file3 := upload2.NewFile()
//	file3.CurrentSize = 3
//
//	err = backend.CreateUpload(ctx, upload2)
//	require.NoError(t, err)
//
//	upload3 := common.NewUpload()
//	upload3.Create()
//	file4 := upload3.NewFile()
//	file4.CurrentSize = 3000
//
//	err = backend.CreateUpload(ctx, upload3)
//	require.NoError(t, err)
//
//	stats, err := backend.GetUserStatistics(ctx, user, token)
//	require.NoError(t, err, "get users error")
//	require.Equal(t, 1, stats.Uploads, "invalid uploads count")
//	require.Equal(t, 1, stats.Files, "invalid files count")
//	require.Equal(t, int64(3), stats.TotalSize, "invalid file size")
//}
