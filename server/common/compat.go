package common

import (
	"encoding/json"
	"fmt"
)

// UploadV1 upload object compatible with Plik <1.3
type UploadV1 struct {
	ID       string `json:"id"`
	Creation int64  `json:"uploadDate"`
	TTL      int    `json:"ttl"`

	DownloadDomain string `json:"downloadDomain"`
	RemoteIP       string `json:"uploadIp,omitempty"`
	Comments       string `json:"comments"`

	Files map[string]*File `json:"files"`

	UploadToken string `json:"uploadToken,omitempty"`
	User        string `json:"user,omitempty"`
	Token       string `json:"token,omitempty"`
	IsAdmin     bool   `json:"admin"`

	Stream    bool `json:"stream"`
	OneShot   bool `json:"oneShot"`
	Removable bool `json:"removable"`

	ProtectedByPassword bool   `json:"protectedByPassword"`
	Login               string `json:"login,omitempty"`
	Password            string `json:"password,omitempty"`

	ProtectedByYubikey bool   `json:"protectedByYubikey"`
	Yubikey            string `json:"yubikey,omitempty"`
}

func UnmarshalUpload(bytes []byte, upload *Upload) (version int, err error) {
	err = json.Unmarshal(bytes, upload)
	if err == nil {
		return 0, nil
	}

	uploadV1 := &UploadV1{}
	err = json.Unmarshal(bytes, uploadV1)
	if err != nil {
		return -1, err
	}

	upload.TTL = uploadV1.TTL
	upload.Comments = uploadV1.Comments

	for _, file := range uploadV1.Files {
		upload.Files = append(upload.Files, file)
	}

	upload.Stream = uploadV1.Stream
	upload.OneShot = uploadV1.OneShot
	upload.Removable = uploadV1.Removable
	upload.Login = uploadV1.Login
	upload.Password = uploadV1.Password

	return 1, nil
}

func MarshalUpload(upload *Upload, version int) (bytes []byte, err error) {
	if version == 0 {
		return json.Marshal(upload)
	}

	if version == 1 {
		uploadV1 := &UploadV1{}

		uploadV1.ID = upload.ID
		uploadV1.Creation = upload.CreatedAt.Unix()
		uploadV1.TTL = upload.TTL
		uploadV1.DownloadDomain = upload.DownloadDomain
		uploadV1.Comments = upload.Comments
		uploadV1.UploadToken = upload.UploadToken

		uploadV1.Stream = upload.Stream
		uploadV1.OneShot = upload.OneShot
		uploadV1.Removable = upload.Removable

		uploadV1.ProtectedByPassword = upload.ProtectedByPassword

		uploadV1.Files = make(map[string]*File)
		for _, file := range upload.Files {
			uploadV1.Files[file.ID] = file
		}

		return json.Marshal(uploadV1)
	}

	return nil, fmt.Errorf("invalid version %d", version)
}
