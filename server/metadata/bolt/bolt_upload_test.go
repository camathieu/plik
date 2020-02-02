package bolt

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
)

func TestBackend_CreateUpload_NoUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.CreateUpload(nil)
	require.Errorf(t, err, "missing upload")
}

func TestBackend_CreateUpload_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("uploads"))
	})
	require.NoError(t, err)

	upload := common.NewUpload()
	upload.Create()

	err = backend.CreateUpload(upload)
	require.Error(t, err)
}

func TestBackend_CreateUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)
}

func TestBackend_CreateUpload_AlreadyExists(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)

	err = backend.CreateUpload(upload)
	require.Error(t, err)
}

func TestBackend_CreateUpload_User(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.User = "user"
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)
}

func TestBackend_CreateUpload_TTL(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.TTL = 86400
	upload.Create()

	err := backend.CreateUpload(upload)
	require.NoError(t, err)
}

func TestBackend_GetUpload_NoUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	_, err := backend.GetUpload("")
	require.Errorf(t, err, "Missing upload")
}

func TestBackend_GetUpload_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("uploads"))
	})
	require.NoError(t, err)

	_, err = backend.GetUpload("id")
	require.Error(t, err)
}

func TestBackend_GetUpload_NotFound(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload, err := backend.GetUpload("id")
	require.NoError(t, err, "error expected")
	require.Nil(t, upload, "no upload expected")
}

func TestBackend_GetUpload_InvalidJSON(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return errors.New("unable to get upload bucket")
		}

		err := bucket.Put([]byte(upload.ID), []byte("invalid_json_value"))
		if err != nil {
			return errors.New("unable to put value")
		}

		return nil
	})
	require.NoError(t, err)

	_, err = backend.GetUpload(upload.ID)
	require.Errorf(t, err, "Unable to unserialize metadata from json")
}

func TestBackend_GetUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return errors.New("unable to get upload bucket")
		}

		jsonValue, err := json.Marshal(upload)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(upload.ID), jsonValue)
		if err != nil {
			return errors.New("unable to put value")
		}

		return nil
	})
	require.NoError(t, err)

	_, err = backend.GetUpload(upload.ID)
	require.NoError(t, err, "unable to get upload")
}

func TestBackend_RemoveUpload_NoUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.RemoveUpload(nil)
	require.Errorf(t, err, "Missing upload")
}

func TestBackend_RemoveUpload_MissingBucket(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("uploads"))
	})
	require.NoError(t, err)

	upload := common.NewUpload()
	upload.Create()

	err = backend.RemoveUpload(upload)
	require.Error(t, err, "missing error")
}

func TestBackend_RemoveUpload_NotFound(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.RemoveUpload(upload)
	require.NoError(t, err, "remove error")
}

func TestBackend_RemoveUpload(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.Create()

	err := backend.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uploads"))
		if bucket == nil {
			return errors.New("unable to get upload bucket")
		}

		jsonValue, err := json.Marshal(upload)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(upload.ID), jsonValue)
		if err != nil {
			return errors.New("unable to put value")
		}

		return nil
	})
	require.NoError(t, err)

	err = backend.RemoveUpload(upload)
	require.NoError(t, err)
}

func TestBackend_RemoveUpload_User(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.User = "user"
	upload.Create()

	err := backend.RemoveUpload(upload)
	require.NoError(t, err)
}

func TestBackend_RemoveUpload_TTL(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	upload := common.NewUpload()
	upload.TTL = 86400
	upload.Create()

	err := backend.RemoveUpload(upload)
	require.NoError(t, err)
}
