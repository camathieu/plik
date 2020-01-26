package common

const FILE_MISSING = "missing"
const FILE_UPLOADING = "uploading"
const FILE_UPLOADED = "uploaded"
const FILE_REMOVED = "removed"
const FILE_DELETED = "deleted"

// File object
type File struct {
	ID             string                 `json:"id" bson:"fileId"`
	Name           string                 `json:"fileName" bson:"fileName"`
	Md5            string                 `json:"fileMd5" bson:"fileMd5"`
	Status         string                 `json:"status" bson:"status"`
	Type           string                 `json:"fileType" bson:"fileType"`
	UploadDate     int64                  `json:"fileUploadDate" bson:"fileUploadDate"`
	CurrentSize    int64                  `json:"fileSize" bson:"fileSize"`
	BackendDetails map[string]interface{} `json:"backendDetails,omitempty" bson:"backendDetails"`
	Reference      string                 `json:"reference" bson:"reference"`
}

// NewFile instantiate a new object
// and generate a random id
func NewFile() (file *File) {
	file = new(File)
	file.ID = GenerateRandomID(16)
	return
}

// GenerateID generate a new File ID
func (file *File) GenerateID() {
	file.ID = GenerateRandomID(16)
}

// Sanitize removes sensible information from
// object. Used to hide information in API.
func (file *File) Sanitize() {
	file.BackendDetails = nil
}
