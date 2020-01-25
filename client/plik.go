/**

    Plik upload client

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/olekukonko/ts"

	"github.com/root-gg/plik/client/archive"
	"github.com/root-gg/plik/client/crypto"
	"github.com/root-gg/plik/plik"
	"github.com/root-gg/utils"
)

// Vars
var arguments map[string]interface{}
var config *CliConfig
var archiveBackend archive.Backend
var cryptoBackend crypto.Backend
var client *plik.Client

var err error

// Main
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	ts.GetSize() // ?

	// Usage /!\ INDENT THIS WITH SPACES NOT TABS /!\
	usage := `plik

Usage:
  plik [options] [FILE] ...

Options:
  -h --help                 Show this help
  -d --debug                Enable debug mode
  -q --quiet                Enable quiet mode
  -o, --oneshot             Enable OneShot ( Each file will be deleted on first download )
  -r, --removable           Enable Removable upload ( Each file can be deleted by anyone at anymoment )
  -S, --stream              Enable Streaming ( It will block until remote user starts downloading )
  -t, --ttl TTL             Time before expiration (Upload will be removed in m|h|d)
  -n, --name NAME           Set file name when piping from STDIN
  --server SERVER           Overrides plik url
  --token TOKEN             Specify an upload token
  --comments COMMENT        Set comments of the upload ( MarkDown compatible )
  -p                        Protect the upload with login and password ( be prompted )
  --password PASSWD         Protect the upload with "login:password" ( if omitted default login is "plik" )
  -y, --yubikey             Protect the upload with a Yubikey OTP
  -a                        Archive upload using default archive params ( see ~/.plikrc )
  --archive MODE            Archive upload using the specified archive backend : tar|zip
  --compress MODE           [tar] Compression codec : gzip|bzip2|xz|lzip|lzma|lzop|compress|no
  --archive-options OPTIONS [tar|zip] Additional command line options
  -s                        Encrypt upload using the default encryption parameters ( see ~/.plikrc )
  --not-secure              Do not encrypt upload files regardless of the ~/.plikrc configurations
  --secure MODE             Encrypt upload files using the specified crypto backend : openssl|pgp
  --cipher CIPHER           [openssl] Openssl cipher to use ( see openssl help )
  --passphrase PASSPHRASE   [openssl] Passphrase or '-' to be prompted for a passphrase
  --recipient RECIPIENT     [pgp] Set recipient for pgp backend ( example : --recipient Bob )
  --secure-options OPTIONS  [openssl|pgp] Additional command line options
  --update                  Update client
  -v --version              Show client version
`
	// Parse command line arguments
	arguments, _ = docopt.Parse(usage, nil, true, "", false)

	// Load config
	config, err = LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load configuration : %s\n", err)
		os.Exit(1)
	}

	// Load arguments
	err = config.UnmarshalArgs(arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if config.Debug {
		fmt.Println("Arguments : ")
		utils.Dump(arguments)
		fmt.Println("Configuration : ")
		utils.Dump(config)
	}

	client = plik.NewClient(config.URL)
	client.Debug = config.Debug
	client.ClientName = "plik_cli"

	// Check client version
	updateFlag := arguments["--update"].(bool)
	err = updateClient(updateFlag)
	if err == nil {
		if updateFlag {
			os.Exit(0)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unable to update Plik client : \n")
		fmt.Fprintf(os.Stderr, "%s\n", err)
		if updateFlag {
			os.Exit(1)
		}
	}

	// Detect STDIN type
	// --> If from pipe : ok, doing nothing
	// --> If not from pipe, and no files in arguments : printing help
	fi, _ := os.Stdin.Stat()

	if runtime.GOOS != "windows" {
		if (fi.Mode()&os.ModeCharDevice) != 0 && len(arguments["FILE"].([]string)) == 0 {
			fmt.Println(usage)
			os.Exit(1)
		}
	} else {
		if len(arguments["FILE"].([]string)) == 0 {
			fmt.Println(usage)
			os.Exit(1)
		}
	}

	upload := client.NewUpload()
	upload.Token = config.Token
	upload.TTL = config.TTL
	upload.Stream = config.Stream
	upload.OneShot = config.OneShot
	upload.Removable = config.Removable
	upload.Comments = config.Comments
	upload.Login = config.Login
	upload.Password = config.Password
	upload.Yubikey = config.yubikeyToken

	if len(config.filePaths) == 0 {
		upload.AddFileFromReader("STDIN", bufio.NewReader(os.Stdin))
	} else {
		if config.Archive {
			archiveBackend, err = archive.NewArchiveBackend(config.ArchiveMethod, config.ArchiveOptions)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to initialize archive backend : %s", err)
				os.Exit(1)
			}

			err = archiveBackend.Configure(arguments)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to configure archive backend : %s", err)
				os.Exit(1)
			}

			reader, err := archiveBackend.Archive(config.filePaths)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create archive : %s", err)
				os.Exit(1)
			}

			filename := archiveBackend.GetFileName(config.filePaths)
			upload.AddFileFromReader(filename, reader)
		} else {
			for _, path := range config.filePaths {
				_, err := upload.AddFileFromPath(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s : %s\n", path, err)
					os.Exit(1)
				}
			}
		}
	}

	if config.filenameOverride != "" {
		if len(upload.Files()) != 1 {
			fmt.Fprintf(os.Stderr, "Can't override filename if more than one file to upload\n")
			os.Exit(1)
		}
		upload.Files()[0].Name = config.filenameOverride
	}

	// Initialize crypto backend
	if config.Secure {
		cryptoBackend, err = crypto.NewCryptoBackend(config.SecureMethod, config.SecureOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to initialize crypto backend : %s", err)
			os.Exit(1)
		}
		err = cryptoBackend.Configure(arguments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to configure crypto backend : %s", err)
			os.Exit(1)
		}
	}

	// Initialize progress bar display
	var progress *Progress
	if !config.Quiet && !config.Debug {
		progress = NewProgress(upload.Files())
	}

	// Add files to upload
	for _, file := range upload.Files() {
		if config.Secure {
			file.WrapReader(func(fileReader io.ReadCloser) io.ReadCloser {
				reader, err := cryptoBackend.Encrypt(fileReader)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to encrypt file :%s", err)
					os.Exit(1)
				}
				return ioutil.NopCloser(reader)
			})
		}

		if !config.Quiet && !config.Debug {
			progress.register(file)
		}
	}

	// Create upload on server
	err = upload.Create()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create upload : %s\n", err)
		os.Exit(1)
	}

	// Mon, 02 Jan 2006 15:04:05 MST
	creationDate := time.Unix(upload.Details().Creation, 0).Format(time.RFC1123)

	// Display upload url
	printf("Upload successfully created at %s : \n", creationDate)

	uploadURL, err := upload.GetURL()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get upload url %s\n", err)
		os.Exit(1)
	} else {
		printf("    %s\n\n", uploadURL)
	}

	if config.Stream && !config.Debug {
		for _, file := range upload.Files() {
			cmd, err := getFileCommand(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to get download command for file %s : %s\n", file.Name, err)
			}
			fmt.Println(cmd)
		}
		printf("\n")
	}

	if !config.Quiet && !config.Debug {
		// Nothing should be printed between this an progress.Stop()
		progress.start()
	}

	// Upload files
	_ = upload.Upload()

	if !config.Quiet && !config.Debug {
		// Finalize the progress bar display
		progress.stop()
	}

	// Display download commands
	if !config.Stream {
		printf("\nCommands : \n")
		for _, file := range upload.Files() {
			// Print file information (only url if quiet mode is enabled)
			if file.Error() != nil {
				continue
			}
			if config.Quiet {
				URL, err := file.GetURL()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get download command for file %s : %s\n", file.Name, err)
				}
				fmt.Println(URL)
			} else {
				cmd, err := getFileCommand(file)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get download command for file %s : %s\n", file.Name, err)
				}
				fmt.Println(cmd)
			}
		}
	} else {
		printf("\n")
	}
}

func getFileCommand(file *plik.File) (command string, err error) {
	// Step one - Downloading file
	switch config.DownloadBinary {
	case "wget":
		command += "wget -q -O-"
	case "curl":
		command += "curl -s"
	default:
		command += config.DownloadBinary
	}

	URL, err := file.GetURL()
	if err != nil {
		return "", err
	}
	command += fmt.Sprintf(` "%s"`, URL)

	// If Ssl
	if config.Secure {
		command += fmt.Sprintf(" | %s", cryptoBackend.Comments())
	}

	// If archive
	if config.Archive {
		if config.ArchiveMethod == "zip" {
			command += fmt.Sprintf(` > '%s'`, file.Name)
		} else {
			command += fmt.Sprintf(" | %s", archiveBackend.Comments())
		}
	} else {
		command += fmt.Sprintf(` > '%s'`, file.Name)
	}

	return
}

func printf(format string, args ...interface{}) {
	if !config.Quiet {
		fmt.Printf(format, args...)
	}
}
