# Plik library

Plik library is a Golang library to upload files to a Plik server.

### Installation

```
go get -v github.com/root-gg/plik/plik
```

#### 1 Easy mode

```go
plik.NewClient("https://plik.server.url")

// To upload regular files
upload, err := client.UploadFiles("/home/file1", "/home/file2", ...)

// To upload a byte stream
upload, err := client.UploadReader("filename", ioReader)
```

#### 2 Full mode

The workflow is :
 - Create a new client
 - Create an Upload
 - Create some Files
 - Add the files to the Upload
 - Create the upload to get the necessary metadata
 - Upload the files

```go
client := plik.NewClient("https://plik.server.url")

// Optional client configuration
client.OneShot = true
client.Token = "xxxx-xxxx-xxxx-xxxx"

upload := client.NewUpload()

// Optional upload configuration
upload.OneShot = false

// Create file from path
file1, err = upload.AddFileFromPath(path)

// Create file from reader
file2, err = upload.AddFileFromReader("filename", ioReader)

// Create upload server side ( optional step that is called by upload.Upload() / file.Upload() if omitted )
err = upload.Create()

// Upload all added files in parallel
err = upload.Upload()

// Upload a single file
err = file.Upload()

// Get upload URL
uploadURL, err := upload.GetURL()

// Get file URL
for _, file := range upload.Files() {
    fileURL, err := file.GetURL()
}
```

#### 3 Bonus

```go
// Get Upload
upload = client.GetUpload(id)

// Download file
reader, err = client.Download(upload.Info(), file.Info())

// Download archive
reader, err = client.DownloadArchive(upload.Info())

// Remove File ( might need to be authenticated )
err = client.RemoveFile(upload.Info(), file.Info())

// Remove Upload ( might need to be authenticated )
err = client.RemoveUpload(upload.Info())

// Get remote server build info
buildInfo, err = client.GetServerVersion()
```