package common

// FileMissing when a file is waiting to be uploaded
const FileMissing = "missing"

// FileUploading when a file is being uploaded
const FileUploading = "uploading"

// FileUploaded when a file has been uploaded and is ready to be downloaded
const FileUploaded = "uploaded"

// FileRemoved when a file has been removed and can't be downloaded anymore but has not yet been deleted
const FileRemoved = "removed"

// FileDeleted when a file has been deleted from the data backend
const FileDeleted = "deleted"

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
