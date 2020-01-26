package common

import (
	"crypto/rand"
	"math/big"
	"time"
)

var (
	randRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

// UploadTx is used to mutate upload metadata
// This must be without side effects as it can be called multiple times to resolve conflicts
type UploadTx func(*Upload) error

// Upload object
type Upload struct {
	ID       string `json:"id" bson:"id"`
	Creation int64  `json:"uploadDate" bson:"uploadDate"`
	TTL      int    `json:"ttl" bson:"ttl"`

	DownloadDomain string `json:"downloadDomain" bson:"-"`
	RemoteIP       string `json:"uploadIp,omitempty" bson:"uploadIp"`
	Comments       string `json:"comments" bson:"comments"`

	Files map[string]*File `json:"files" bson:"files"`

	UploadToken string `json:"uploadToken,omitempty" bson:"uploadToken"`
	User        string `json:"user,omitempty" bson:"user"`
	Token       string `json:"token,omitempty" bson:"token"`
	Admin       bool   `json:"admin"`

	Stream    bool `json:"stream" bson:"stream"`
	OneShot   bool `json:"oneShot" bson:"oneShot"`
	Removable bool `json:"removable" bson:"removable"`

	ProtectedByPassword bool   `json:"protectedByPassword" bson:"protectedByPassword"`
	Login               string `json:"login,omitempty" bson:"login"`
	Password            string `json:"password,omitempty" bson:"password"`

	ProtectedByYubikey bool   `json:"protectedByYubikey" bson:"protectedByYubikey"`
	Yubikey            string `json:"yubikey,omitempty" bson:"yubikey"`
}

// NewUpload instantiate a new upload object
func NewUpload() (upload *Upload) {
	upload = new(Upload)
	upload.Files = make(map[string]*File)
	return
}

// Create fills token, id, date
// We have split in two functions because, the unmarshalling made
// in http handlers would erase the fields
func (upload *Upload) Create() {
	upload.ID = GenerateRandomID(16)
	upload.Creation = time.Now().Unix()
	if upload.Files == nil {
		upload.Files = make(map[string]*File)
	}
	upload.UploadToken = GenerateRandomID(32)
}

// NewFile creates a new file and add it to the current upload
func (upload *Upload) NewFile() (file *File) {
	file = NewFile()
	upload.Files[file.ID] = file
	return file
}

// Sanitize removes sensible information from
// object. Used to hide information in API.
func (upload *Upload) Sanitize() {
	upload.RemoteIP = ""
	upload.Login = ""
	upload.Password = ""
	upload.UploadToken = ""
	upload.User = ""
	upload.Token = ""
	upload.Yubikey = ""
	for _, file := range upload.Files {
		file.Sanitize()
	}
}

// GenerateRandomID generates a random string with specified length.
// Used to generate upload id, tokens, ...
func GenerateRandomID(length int) string {
	max := *big.NewInt(int64(len(randRunes)))
	b := make([]rune, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, &max)
		b[i] = randRunes[n.Int64()]
	}

	return string(b)
}

// IsExpired check if the upload is expired
func (upload *Upload) IsExpired() bool {
	if upload.TTL > 0 {
		if time.Now().Unix() >= (upload.Creation + int64(upload.TTL)) {
			return true
		}
	}
	return false
}
