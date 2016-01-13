/* The MIT License (MIT)

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
THE SOFTWARE. */

package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/facebookgo/httpdown"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/dataBackend"
	"github.com/root-gg/plik/server/handlers"
	"github.com/root-gg/plik/server/metadataBackend"
	"github.com/root-gg/plik/server/middleware"
	"github.com/root-gg/plik/server/shortenBackend"
)

var log *logger.Logger

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log = common.Logger()

	var configFile = flag.String("config", "plikd.cfg", "Configuration file (default: plikd.cfg")
	var version = flag.Bool("version", false, "Show version of plikd")
	var port = flag.Int("port", 0, "Overrides plik listen port")
	flag.Parse()
	if *version {
		fmt.Printf("Plik server %s\n", common.GetBuildInfo())
		os.Exit(0)
	}

	common.LoadConfiguration(*configFile)
	log.Infof("Starting plikd server v" + common.GetBuildInfo().Version)

	// Overrides port if provided in command line
	if *port != 0 {
		common.Config.ListenPort = *port
	}

	// Initialize all backends
	metadataBackend.Initialize()
	dataBackend.Initialize()
	shortenBackend.Initialize()

	// Initialize the httpdown wrapper
	hd := &httpdown.HTTP{
		StopTimeout: 5 * time.Minute,
		KillTimeout: 1 * time.Second,
	}

	stdChain := juliet.NewChain(middleware.Logger, middleware.SourceIP, middleware.Log, middleware.Authenticate)
	stdChainWithRedirect := juliet.NewChain(middleware.RedirectOnFailure).AppendChain(stdChain)

	// HTTP Api routes configuration
	r := mux.NewRouter()
	r.Handle("/config", stdChain.Then(handlers.GetConfiguration)).Methods("GET")
	r.Handle("/version", stdChain.Then(handlers.GetVersion)).Methods("GET")
	r.Handle("/upload", stdChain.Then(handlers.CreateUpload)).Methods("POST")
	r.Handle("/upload/{uploadID}", stdChain.Append(middleware.Upload).Then(handlers.GetUpload)).Methods("GET")
	r.Handle("/upload/{uploadID}", stdChain.Append(middleware.Upload).Then(handlers.RemoveUpload)).Methods("DELETE")
	r.Handle("/file/{uploadID}", stdChain.Append(middleware.Upload).Then(handlers.AddFile)).Methods("POST")
	r.Handle("/file/{uploadID}/{fileID}/{filename}", stdChain.Append(middleware.Upload, middleware.File).Then(handlers.AddFile)).Methods("POST")
	r.Handle("/file/{uploadID}/{fileID}/{filename}", stdChain.Append(middleware.Upload, middleware.File).Then(handlers.RemoveFile)).Methods("DELETE")
	r.Handle("/file/{uploadID}/{fileID}/{filename}", stdChainWithRedirect.Append(middleware.Upload, middleware.File).Then(handlers.GetFile)).Methods("HEAD", "GET")
	r.Handle("/file/{uploadID}/{fileID}/{filename}/yubikey/{yubikey}", stdChainWithRedirect.Then(handlers.GetFile)).Methods("GET")
	r.Handle("/stream/{uploadID}/{fileID}/{filename}", stdChain.Append(middleware.Upload, middleware.File).Then(handlers.AddFile)).Methods("POST")
	r.Handle("/stream/{uploadID}/{fileID}/{filename}", stdChain.Append(middleware.Upload, middleware.File).Then(handlers.RemoveFile)).Methods("DELETE")
	r.Handle("/stream/{uploadID}/{fileID}/{filename}", stdChainWithRedirect.Append(middleware.Upload, middleware.File).Then(handlers.GetFile)).Methods("HEAD", "GET")
	r.Handle("/stream/{uploadID}/{fileID}/{filename}/yubikey/{yubikey}", stdChainWithRedirect.Then(handlers.GetFile)).Methods("GET")
	r.Handle("/auth/google/login", stdChain.Then(handlers.GoogleLogin)).Methods("GET")
	r.Handle("/auth/google/callback", stdChainWithRedirect.Then(handlers.GoogleCallback)).Methods("GET")
	r.Handle("/auth/logout", stdChain.Then(handlers.Logout)).Methods("GET")
	r.Handle("/me", stdChain.Then(handlers.UserInfo)).Methods("GET")
	r.Handle("/me", stdChain.Then(handlers.DeleteAccount)).Methods("DELETE")
	//  r.Handle("/me/token", stdChain.Then(handlers.GetUploadsWithTokenHandler)).Methods("POST")
	//  r.Handle("/me/token/{token}", stdChain.Then(handlers.RevokeTokenHandler)).Methods("DELETE")
	r.Handle("/me/uploads", stdChain.Then(handlers.GetUserUploads)).Methods("GET")
	r.Handle("/me/uploads", stdChain.Then(handlers.RemoveUserUploads)).Methods("DELETE")
	r.Handle("/qrcode", stdChain.Then(handlers.GetQrCode)).Methods("GET")
	r.PathPrefix("/clients/").Handler(http.StripPrefix("/clients/", http.FileServer(http.Dir("../clients"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.Handle("/", r)

	go UploadsCleaningRoutine()

	// Start HTTP server
	var err error
	var server *http.Server

	address := common.Config.ListenAddress + ":" + strconv.Itoa(common.Config.ListenPort)

	if common.Config.SslEnabled {

		// Load cert
		cert, err := tls.LoadX509KeyPair(common.Config.SslCert, common.Config.SslKey)
		if err != nil {
			log.Fatalf("Unable to load ssl certificate : %s", err)
		}

		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS10, Certificates: []tls.Certificate{cert}}
		server = &http.Server{Addr: address, Handler: r, TLSConfig: tlsConfig}
	} else {
		server = &http.Server{Addr: address, Handler: r}
	}

	err = httpdown.ListenAndServe(server, hd)
	if err != nil {
		log.Fatalf("Unable to start HTTP server : %s", err)
	}

}

//
//// Misc functions
//

// UploadsCleaningRoutine periodicaly remove expired uploads
func UploadsCleaningRoutine() {
	ctx := juliet.NewContext()

	for {
		// Sleep between 2 hours and 3 hours
		// This is a dirty trick to avoid frontends doing this at the same time
		r, _ := rand.Int(rand.Reader, big.NewInt(3600))
		randomSleep := r.Int64() + 7200

		log.Infof("Will clean old uploads in %d seconds.", randomSleep)
		time.Sleep(time.Duration(randomSleep) * time.Second)
		log.Infof("Cleaning expired uploads...")

		// Get uploads that needs to be removed
		uploadIds, err := metadataBackend.GetMetaDataBackend().GetUploadsToRemove(ctx)
		if err != nil {
			log.Warningf("Failed to get expired uploads : %s", err)
		} else {
			// Remove them
			for _, uploadID := range uploadIds {
				log.Infof("Removing expired upload %s", uploadID)
				// Get upload metadata
				upload, err := metadataBackend.GetMetaDataBackend().Get(ctx, uploadID)
				if err != nil {
					log.Warningf("Unable to get infos for upload: %s", err)
					continue
				}

				// Remove from data backend
				err = dataBackend.GetDataBackend().RemoveUpload(ctx, upload)
				if err != nil {
					log.Warningf("Unable to remove upload data : %s", err)
					continue
				}

				// Remove from metadata backend
				err = metadataBackend.GetMetaDataBackend().Remove(ctx, upload)
				if err != nil {
					log.Warningf("Unable to remove upload metadata : %s", err)
				}
			}
		}
	}
}
