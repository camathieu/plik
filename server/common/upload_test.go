package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUpload(t *testing.T) {
	upload := NewUpload()
	require.NotNil(t, upload, "invalid upload")
	require.NotNil(t, upload.Files, "invalid upload")
}

func TestUploadCreate(t *testing.T) {
	upload := NewUpload()
	upload.Create()
	require.NotZero(t, upload.ID, "missing id")
	require.NotZero(t, upload.Creation, "missing creation date")
}

func TestUploadNewFile(t *testing.T) {
	upload := NewUpload()
	file := upload.NewFile()
	require.NotZero(t, len(upload.Files), "invalid file count")
	require.Equal(t, file, upload.Files[file.ID], "missing file")
}

func TestUploadSanitize(t *testing.T) {
	upload := NewUpload()
	upload.RemoteIP = "ip"
	upload.Login = "login"
	upload.Password = "password"
	upload.Yubikey = "token"
	upload.UploadToken = "token"
	upload.Token = "token"
	upload.User = "user"
	upload.Sanitize()

	require.Zero(t, upload.RemoteIP, "invalid sanitized upload")
	require.Zero(t, upload.Login, "invalid sanitized upload")
	require.Zero(t, upload.Password, "invalid sanitized upload")
	require.Zero(t, upload.Yubikey, "invalid sanitized upload")
	require.Zero(t, upload.UploadToken, "invalid sanitized upload")
	require.Zero(t, upload.Token, "invalid sanitized upload")
	require.Zero(t, upload.UploadToken, "invalid sanitized upload")
}
