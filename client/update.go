package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/utils"
)

func updateClient(updateFlag bool) (err error) {
	// Do not check for update if AutoUpdate is not enabled
	if !updateFlag && !config.AutoUpdate {
		return
	}

	// Do not update when quiet mode is enabled
	if !updateFlag && config.Quiet {
		return
	}

	// Get client MD5SUM
	path, err := osext.Executable()
	if err != nil {
		return
	}
	currentMD5, err := utils.FileMd5sum(path)
	if err != nil {
		return
	}

	// Check server version
	currentVersion := common.GetBuildInfo().Version

	var newVersion string
	var downloadURL string
	var newMD5 string
	var buildInfo *common.BuildInfo

	var URL *url.URL
	URL, err = url.Parse(config.URL + "/version")
	if err != nil {
		err = fmt.Errorf("Unable to get server version : %s", err)
		return
	}
	var req *http.Request
	req, err = http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		err = fmt.Errorf("Unable to get server version : %s", err)
		return
	}

	resp, err := client.MakeRequest(req)
	if resp == nil {
		err = fmt.Errorf("Unable to get server version : %s", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// >=1.1 use BuildInfo from /version

		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("Unable to get server version : %s", err)
			return
		}

		// Parse json BuildInfo object
		buildInfo = new(common.BuildInfo)
		err = json.Unmarshal(body, buildInfo)
		if err != nil {
			err = fmt.Errorf("Unable to get server version : %s", err)
			return
		}

		newVersion = buildInfo.Version
		for _, client := range buildInfo.Clients {
			if client.OS == runtime.GOOS && client.ARCH == runtime.GOARCH {
				newMD5 = client.Md5
				downloadURL = config.URL + "/" + client.Path
				break
			}
		}

		if newMD5 == "" || downloadURL == "" {
			err = fmt.Errorf("Server does not offer a %s-%s client", runtime.GOOS, runtime.GOARCH)
			return
		}
	} else if resp.StatusCode == 404 {
		// <1.1 fallback on MD5SUM file

		baseURL := config.URL + "/clients/" + runtime.GOOS + "-" + runtime.GOARCH
		var URL *url.URL
		URL, err = url.Parse(baseURL + "/MD5SUM")
		if err != nil {
			return
		}
		var req *http.Request
		req, err = http.NewRequest("GET", URL.String(), nil)
		if err != nil {
			err = fmt.Errorf("Unable to get server version : %s", err)
			return
		}

		resp, err = client.MakeRequest(req)
		if err != nil {
			err = fmt.Errorf("Unable to get server version : %s", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = fmt.Errorf("Unable to get server version : %s", resp.Status)
			return
		}

		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("Unable to get server version : %s", err)
			return
		}
		newMD5 = utils.Chomp(string(body))

		binary := "plik"
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}
		downloadURL = baseURL + "/" + binary
	} else {
		err = fmt.Errorf("Unable to get server version : %s", err)
		return
	}

	// Check if the client is up to date
	if currentMD5 == newMD5 {
		if updateFlag {
			if newVersion != "" {
				printf("Plik client %s is up to date\n", newVersion)
			} else {
				printf("Plik client is up to date\n")
			}
			os.Exit(0)
		}
		return
	}

	// Ask for permission
	if newVersion != "" {
		fmt.Printf("Update Plik client from %s to %s ? [Y/n] ", currentVersion, newVersion)
	} else {
		fmt.Printf("Update Plik client to match server version ? [Y/n] ")
	}
	input := "y"
	fmt.Scanln(&input)
	if !strings.HasPrefix(strings.ToLower(input), "y") {
		if updateFlag {
			os.Exit(0)
		}
		return
	}

	// Display release notes
	if buildInfo != nil && buildInfo.Releases != nil {

		// Find current release
		currentReleaseIndex := -1
		for i, release := range buildInfo.Releases {
			if release.Name == currentVersion {
				currentReleaseIndex = i
			}
		}

		// Find new release
		newReleaseIndex := -1
		for i, release := range buildInfo.Releases {
			if release.Name == newVersion {
				newReleaseIndex = i
			}
		}

		// Find releases between current and new version
		var releases []*common.Release
		if currentReleaseIndex > 0 && newReleaseIndex > 0 && currentReleaseIndex < newReleaseIndex {
			releases = buildInfo.Releases[currentReleaseIndex+1 : newReleaseIndex+1]
		}

		for _, release := range releases {
			// Get release notes from server
			var URL *url.URL
			URL, err = url.Parse(config.URL + "/changelog/" + release.Name)
			if err != nil {
				continue
			}
			var req *http.Request
			req, err = http.NewRequest("GET", URL.String(), nil)
			if err != nil {
				err = fmt.Errorf("Unable to get release notes for version %s : %s", release.Name, err)
				continue
			}

			resp, err = client.MakeRequest(req)
			if err != nil {
				err = fmt.Errorf("Unable to get release notes for version %s : %s", release.Name, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				err = fmt.Errorf("Unable to get release notes for version %s : %s", release.Name, err)
				continue
			}

			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				err = fmt.Errorf("Unable to get release notes for version %s : %s", release.Name, err)
				continue
			}

			// Ask to display the release notes
			fmt.Printf("Do you want to browse the release notes of version %s ? [Y/n] ", release.Name)
			input := "y"
			fmt.Scanln(&input)
			if !strings.HasPrefix(strings.ToLower(input), "y") {
				continue
			}

			// Display the release notes
			releaseDate := time.Unix(release.Date, 0).Format("Mon Jan 2 2006 15:04")
			fmt.Printf("Plik %s has been released %s\n\n", release.Name, releaseDate)
			fmt.Println(string(body))

			// Let user review the last release notes and ask to confirm update
			if release.Name == newVersion {
				fmt.Printf("\nUpdate Plik client from %s to %s ? [Y/n] ", currentVersion, newVersion)
				input = "y"
				fmt.Scanln(&input)
				if !strings.HasPrefix(strings.ToLower(input), "y") {
					if updateFlag {
						os.Exit(0)
					}
					return
				}
				break
			}
		}
	}

	// Download new client
	tmpPath := filepath.Dir(path) + "/" + "." + filepath.Base(path) + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	URL, err = url.Parse(downloadURL)
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}
	req, err = http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}
	resp, err = client.MakeRequest(req)
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Unable to download client : %s", resp.Status)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}
	err = tmpFile.Close()
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}

	// Check download integrity
	downloadMD5, err := utils.FileMd5sum(tmpPath)
	if err != nil {
		err = fmt.Errorf("Unable to download client : %s", err)
		return
	}
	if downloadMD5 != newMD5 {
		err = fmt.Errorf("Unable to download client : md5sum %s does not match %s", downloadMD5, newMD5)
		return
	}

	// Replace old client
	err = os.Rename(tmpPath, path)
	if err != nil {
		err = fmt.Errorf("Unable to replace client : %s", err)
		return
	}

	if newVersion != "" {
		fmt.Printf("Plik client successfully updated to %s\n", newVersion)
	} else {
		fmt.Printf("Plik client successfully updated\n")
	}

	return
}
