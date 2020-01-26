package weedfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/root-gg/utils"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
)

var (
	client = http.Client{}
)

// Ensure WeedFS Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

// Config describes configuration for WeedFS data backend
type Config struct {
	MasterURL          string
	ReplicationPattern string
}

// NewConfig instantiate a new default configuration
// and override it with configuration passed as argument
func NewConfig(params map[string]interface{}) (config *Config) {
	config = new(Config)
	config.MasterURL = "http://127.0.0.1:9333"
	config.ReplicationPattern = "000"
	utils.Assign(config, params)
	return
}

// Backend object
type Backend struct {
	Config *Config
}

// NewBackend instantiate a new WeedFS Data Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend) {
	b = new(Backend)
	b.Config = config
	return
}

// GetFile implementation for WeedFS Data Backend
func (b *Backend) GetFile(upload *common.Upload, id string) (reader io.ReadCloser, err error) {
	file := upload.Files[id]

	// Get WeedFS volume from upload metadata
	if file.BackendDetails["WeedFsVolume"] == nil {
		return nil, fmt.Errorf("missing WeedFS volume from backend details")
	}
	weedFsVolume := file.BackendDetails["WeedFsVolume"].(string)

	// Get WeedFS file id from upload metadata
	if file.BackendDetails["WeedFsFileID"] == nil {
		return nil, fmt.Errorf("missing WeedFS file id from backend details")
	}
	WeedFsFileID := file.BackendDetails["WeedFsFileID"].(string)

	// Get WeedFS volume url
	volumeURL, err := b.getvolumeURL(weedFsVolume)
	if err != nil {
		return nil, fmt.Errorf("unable to get WeedFS volume url %s", weedFsVolume)
	}

	// Get file from WeedFS volume, the response will be
	// piped directly to the client response body
	fileCompleteURL := "http://" + volumeURL + "/" + weedFsVolume + "," + WeedFsFileID
	resp, err := http.Get(fileCompleteURL)
	if err != nil {
		return nil, fmt.Errorf("error while downloading file from WeedFS at %s : %s", fileCompleteURL, err)
	}

	return resp.Body, nil
}

// AddFile implementation for WeedFS Data Backend
func (b *Backend) AddFile(upload *common.Upload, file *common.File, fileReader io.Reader) (backendDetails map[string]interface{}, err error) {
	backendDetails = make(map[string]interface{})

	// Request a volume and a new file id from a WeedFS master
	assignURL := b.Config.MasterURL + "/dir/assign?replication=" + b.Config.ReplicationPattern

	resp, err := client.Post(assignURL, "", nil)
	if err != nil {
		return nil, fmt.Errorf("error while getting id from WeedFS master at %s : %s", assignURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body from WeedFS master at %s : %s", assignURL, err)
	}

	// Unserialize response body
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(bodyStr, &responseMap)
	if err != nil {
		return nil, fmt.Errorf("unable to unserialize json response \"%s\" from WeedFS master at %s : %s", bodyStr, assignURL, err)
	}

	if responseMap["fid"] != nil && responseMap["fid"].(string) != "" {
		splitVolumeFromID := strings.Split(responseMap["fid"].(string), ",")
		if len(splitVolumeFromID) > 1 {
			backendDetails["WeedFsVolume"] = splitVolumeFromID[0]
			backendDetails["WeedFsFileID"] = splitVolumeFromID[1]
		} else {
			return nil, fmt.Errorf("invalid fid from WeedFS master response \"%s\" at %s", bodyStr, assignURL)
		}
	} else {
		return nil, fmt.Errorf("missing fid from WeedFS master response \"%s\" at %s", bodyStr, assignURL)
	}

	// Construct upload url
	if responseMap["publicUrl"] == nil || responseMap["publicUrl"].(string) == "" {
		return nil, fmt.Errorf("missing publicUrl from WeedFS master response \"%s\" at %s", bodyStr, assignURL)
	}
	fileURL := "http://" + responseMap["publicUrl"].(string) + "/" + responseMap["fid"].(string)
	var URL *url.URL
	URL, err = url.Parse(fileURL)
	if err != nil {
		return nil, fmt.Errorf("unable to construct WeedFS upload url \"%s\"", fileURL)
	}

	// Pipe the uploaded file from the client request body
	// to the WeedFS request body without buffering
	pipeReader, pipeWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(pipeWriter)
	go func() {
		filePart, err := multipartWriter.CreateFormFile("file", file.Name)
		if err != nil {
			err = fmt.Errorf("unable to create multipart form : %s", err)
			_ = pipeWriter.CloseWithError(err)
			return
		}

		_, err = io.Copy(filePart, fileReader)
		if err != nil {
			err = fmt.Errorf("unable to copy file to WeedFS request body : %s", err)
			_ = pipeWriter.CloseWithError(err)
			return
		}

		err = multipartWriter.Close()
		if err != nil {
			err = fmt.Errorf("unable to close multipartWriter : %s", err)
			_ = pipeWriter.CloseWithError(err)
			return
		}

		_ = pipeWriter.Close()
	}()

	// Upload file to WeedFS volume
	req, err := http.NewRequest("PUT", URL.String(), pipeReader)
	if err != nil {
		return nil, fmt.Errorf("unable to create PUT request to %s : %s", URL.String(), err)
	}
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())
	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to upload file to WeedFS at %s : %s", URL.String(), err)
	}
	defer resp.Body.Close()

	return backendDetails, nil
}

// RemoveFile implementation for WeedFS Data Backend
func (b *Backend) RemoveFile(upload *common.Upload, id string) (err error) {
	// Get file metadata
	file := upload.Files[id]

	// Get WeedFS volume and file id from upload metadata
	if file.BackendDetails["WeedFsVolume"] == nil {
		return fmt.Errorf("missing WeedFS volume from backend details")
	}
	weedFsVolume := file.BackendDetails["WeedFsVolume"].(string)

	if file.BackendDetails["WeedFsFileID"] == nil {
		return fmt.Errorf("missing WeedFS file id from backend details")
	}
	WeedFsFileID := file.BackendDetails["WeedFsFileID"].(string)

	// Get the WeedFS volume url
	volumeURL, err := b.getvolumeURL(weedFsVolume)
	if err != nil {
		return err
	}

	// Construct Url
	fileURL := "http://" + volumeURL + "/" + weedFsVolume + "," + WeedFsFileID
	var URL *url.URL
	URL, err = url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("unable to construct WeedFS url \"%s\"", fileURL)
	}

	// Remove file from WeedFS volume
	req, err := http.NewRequest("DELETE", URL.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create DELETE request to %s : %s", URL.String(), err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to delete file from WeedFS volume at %s : %s", URL.String(), err)
	}
	_ = resp.Body.Close()

	return
}

// RemoveUpload implementation for WeedFS Data Backend
// Iterates on every file and call removeFile
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	for fileID := range upload.Files {
		err = b.RemoveFile(upload, fileID)
		if err != nil {
			return
		}
	}

	return nil
}

func (b *Backend) getvolumeURL(volumeID string) (URL string, err error) {
	// Ask a WeedFS master the volume urls
	URL = b.Config.MasterURL + "/dir/lookup?volumeId=" + volumeID
	resp, err := client.Post(URL, "", nil)
	if err != nil {
		return "", fmt.Errorf("unable to get volume %s url from WeedFS master at %s : %s", volumeID, URL, err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response from WeedFS master at %s : %s", URL, err)
	}

	// Unserialize response body
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(bodyStr, &responseMap)
	if err != nil {
		return "", fmt.Errorf("unable to unserialize json response \"%s\"from WeedFS master at %s : %s", bodyStr, URL, err)
	}

	// As volumes can be replicated there may be more than one
	// available url for a given volume
	var urlsFound []string
	if responseMap["locations"] == nil {
		return "", fmt.Errorf("missing url from WeedFS master response \"%s\" at %s", bodyStr, URL)
	}
	if locationsArray, ok := responseMap["locations"].([]interface{}); ok {
		for _, location := range locationsArray {
			if locationInfos, ok := location.(map[string]interface{}); ok {
				if locationInfos["publicUrl"] != nil {
					if foundURL, ok := locationInfos["publicUrl"].(string); ok {
						urlsFound = append(urlsFound, foundURL)
					}
				}
			}
		}
	}
	if len(urlsFound) == 0 {
		return "", fmt.Errorf("no url found for WeedFS volume %s", volumeID)
	}

	// Take a random url from the list
	URL = urlsFound[rand.Intn(len(urlsFound))]
	return URL, nil
}
