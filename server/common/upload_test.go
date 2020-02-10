package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadCreate(t *testing.T) {
	upload := &Upload{}
	upload.PrepareInsertForTests()
	require.NotZero(t, upload.ID, "missing id")
}

func TestUploadNewFile(t *testing.T) {
	upload := &Upload{}
	upload.NewFile()
	require.NotZero(t, len(upload.Files), "invalid file count")
}

func TestUploadSanitize(t *testing.T) {
	upload := &Upload{}
	upload.RemoteIP = "ip"
	upload.Login = "login"
	upload.Password = "password"
	upload.UploadToken = "token"
	upload.Token = "token"
	upload.User = "user"
	upload.Sanitize()

	require.Zero(t, upload.RemoteIP, "invalid sanitized upload")
	require.Zero(t, upload.Login, "invalid sanitized upload")
	require.Zero(t, upload.Password, "invalid sanitized upload")
	require.Zero(t, upload.UploadToken, "invalid sanitized upload")
	require.Zero(t, upload.Token, "invalid sanitized upload")
	require.Zero(t, upload.UploadToken, "invalid sanitized upload")
}
