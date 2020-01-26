package bolt

//import (
//	"testing"
//	"time"
//
//	"github.com/boltdb/bolt"
//	"github.com/root-gg/plik/server/common"
//	"github.com/root-gg/plik/server/context"
//	"github.com/stretchr/testify/require"
//)
//
//func TestBackend_GetUploadsToRemove_MissingBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("uploads"))
//	})
//	require.NoError(t, err, "unable to remove uploads bucket")
//
//	_, err = backend.GetUploadsToRemove(ctx)
//	require.Error(t, err)
//}
//
//func TestBackend_GetUploadsToRemove(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	upload := common.NewUpload()
//	upload.Create()
//	upload.TTL = 1
//	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()
//
//	err := backend.CreateUpload(ctx, upload)
//	require.NoError(t, err)
//
//	upload2 := common.NewUpload()
//	upload2.Create()
//	upload.TTL = 0
//	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()
//
//	err = backend.CreateUpload(ctx, upload2)
//	require.NoError(t, err)
//
//	upload3 := common.NewUpload()
//	upload3.Create()
//	upload.TTL = 86400
//	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()
//
//	err = backend.CreateUpload(ctx, upload3)
//	require.NoError(t, err)
//
//	ids, err := backend.GetUploadsToRemove(ctx)
//	require.NoError(t, err, "get upload to remove error")
//	require.Len(t, ids, 1, "invalid uploads to remove count")
//}
//
//func TestBackend_GetServerStatistics_MissingBucket(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	err := backend.db.Update(func(tx *bolt.Tx) error {
//		return tx.DeleteBucket([]byte("uploads"))
//	})
//	require.NoError(t, err, "unable to remove uploads bucket")
//
//	_, err = backend.GetServerStatistics(ctx)
//	require.Error(t, err, "missing error")
//}
//
//func TestBackend_GetServerStatistics(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//
//	type pair struct {
//		typ   string
//		size  int64
//		count int
//	}
//
//	plan := []pair{
//		{"type1", 1, 1},
//		{"type2", 1000, 5},
//		{"type3", 1000 * 1000, 10},
//		{"type4", 1000 * 1000 * 1000, 15},
//	}
//
//	for _, item := range plan {
//		for i := 0; i < item.count; i++ {
//			upload := common.NewUpload()
//			upload.Create()
//			file := upload.NewFile()
//			file.Type = item.typ
//			file.CurrentSize = item.size
//
//			err := backend.CreateUpload(ctx, upload)
//			require.NoError(t, err)
//		}
//	}
//
//	stats, err := backend.GetServerStatistics(ctx)
//	require.NoError(t, err, "get server statistics error")
//	require.NotNil(t, stats, "invalid server statistics")
//	require.Equal(t, 31, stats.Uploads, "invalid upload count")
//	require.Equal(t, 31, stats.Files, "invalid files count")
//	require.Equal(t, int64(15010005001), stats.TotalSize, "invalid total file size")
//	require.Equal(t, 31, stats.AnonymousUploads, "invalid anonymous upload count")
//	require.Equal(t, int64(15010005001), stats.AnonymousSize, "invalid anonymous total file size")
//	require.Equal(t, 10, len(stats.FileTypeByCount), "invalid file type by count length")
//	require.Equal(t, "type4", stats.FileTypeByCount[0].Type, "invalid file type by count type")
//	require.Equal(t, 10, len(stats.FileTypeBySize), "invalid file type by size length")
//	require.Equal(t, "type4", stats.FileTypeBySize[0].Type, "invalid file type by size type")
//
//}
