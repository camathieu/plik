package plik

import (
	"bytes"
	"github.com/root-gg/plik/server/common"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const TOKEN = "22b2c7f9-dead-dead-dead-ee8edd115e8a"

func defaultUser() *common.User {
	return &common.User{
		ID:     "ovh:gg1-ovh",
		Name:   "plik",
		Email:  "plik@root.gg",
		Tokens: []*common.Token{{Token: TOKEN}},
	}
}

func TestTokenAuthentication(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	ps.GetConfig().Authentication = true
	ps.GetConfig().NoAnonymousUploads = true

	user := common.NewUser()
	user.ID = "ovh:gg1-ovh"
	t1 := user.NewToken()

	err := start(ps)
	require.NoError(t, err, "unable to start Plik server")

	err = ps.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to create user")

	// Set token for client
	pc.Token = t1.Token

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", ioutil.NopCloser(bytes.NewBufferString(data)))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Metadata().Files, 0, "invalid file count")

	reader, err := file.Download()
	require.NoError(t, err, "unable to download file")
	content, err := ioutil.ReadAll(reader)
	require.NoError(t, err, "unable to read file")
	require.Equal(t, data, string(content), "invalid file content")
}

// A user authenticated with a token should not be able to control an upload authenticated with another token
func TestTokenMultipleToken(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	ps.GetConfig().Authentication = true
	ps.GetConfig().NoAnonymousUploads = true

	err := start(ps)
	require.NoError(t, err, "unable to start Plik server")

	user := common.NewUser()
	user.ID = "ovh:gg2-ovh"
	t1 := user.NewToken()
	t2 := user.NewToken()

	err = ps.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to create user")

	upload := pc.NewUpload()

	// Set token for upload
	upload.Token = t1.Token

	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	require.NoError(t, err, "unable to upload")

	upload.Metadata().UploadToken = ""

	// try to add file to upload with the good token
	upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")

	upload.Token = t2.Token

	// try to add file to upload with the wrong token
	f2 := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	common.RequireError(t, err, "failed to upload at least one file")
	common.RequireError(t, f2.Error(), "you are not allowed to add file to this upload")

	// try to remove file to upload with the wrong token
	err = file.Delete()
	common.RequireError(t, err, "you are not allowed to remove files from this upload")

	// try to remove upload with the wrong token
	err = upload.Delete()
	common.RequireError(t, err, "you are not allowed to remove this upload")

	upload.Token = t1.Token

	// try to remove file with the good token
	err = file.Delete()
	require.NoError(t, err, "unable to remove file")
}

// An admin user authenticated with a token should not have more power than a classical user authenticated with a token
// This is to lower the impact of the leak of an Admin user token
func TestTokenMultipleTokenAdmin(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer shutdown(ps)

	uid := "ovh:gg3-ovh"
	ps.GetConfig().Authentication = true
	ps.GetConfig().NoAnonymousUploads = true
	ps.GetConfig().Admins = append(ps.GetConfig().Admins, uid)

	err := start(ps)
	require.NoError(t, err, "unable to start Plik server")

	user := common.NewUser()
	user.ID = uid
	t1 := user.NewToken()
	t2 := user.NewToken()

	err = ps.GetMetadataBackend().CreateUser(user)
	require.NoError(t, err, "unable to create user")

	upload := pc.NewUpload()
	upload.Token = t1.Token
	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	require.NoError(t, err, "unable to upload")

	upload.Metadata().UploadToken = ""

	// try to add file to upload with the good token
	upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	require.NoError(t, err, "unable to upload file")

	upload.Token = t2.Token

	// try to add file to upload with the wrong token
	f2 := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))
	err = upload.Upload()
	common.RequireError(t, err, "failed to upload at least one file")
	common.RequireError(t, f2.Error(), "you are not allowed to add file to this upload")

	// try to remove file to upload with the wrong token
	err = file.Delete()
	common.RequireError(t, err, "you are not allowed to remove files from this upload")

	// try to remove upload with the wrong token
	err = upload.Delete()
	common.RequireError(t, err, "you are not allowed to remove this upload")

	upload.Token = t1.Token

	// try to remove file with the good token
	err = file.Delete()
	require.NoError(t, err, "Unable to remove file")
}
