/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015> Copyright holders list can be found in AUTHORS file
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

package server

import (
	"errors"
	"testing"
	"time"

	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
	"github.com/stretchr/testify/require"
)

func newPlikServer() (ps *PlikServer) {
	ps = NewPlikServer(common.NewConfiguration())
	ps.config.ListenAddress = "127.0.0.1"
	ps.config.ListenPort = common.APIMockServerDefaultPort
	ps.config.AutoClean(false)
	ps.WithMetadataBackend(metadata_test.NewBackend())
	ps.WithDataBackend(data_test.NewBackend())
	ps.WithStreamBackend(data_test.NewBackend())
	return ps
}

func TestNewPlikServer(t *testing.T) {
	config := common.NewConfiguration()
	ps := NewPlikServer(config)
	require.NotNil(t, ps, "invalid nil Plik server")
	require.Equal(t, logger.INFO, ps.logger.MinLevel, "invalid logger level")
	require.NotNil(t, ps.GetConfig(), "invalid nil configuration")

	config.LogLevel = "DEBUG"
	ps2 := NewPlikServer(config)
	require.NotNil(t, ps2, "invalid nil Plik server")
	require.Equal(t, logger.DEBUG, ps2.logger.MinLevel, "invalid logger level")
}

func TestStartShutdownPlikServer(t *testing.T) {
	ps := newPlikServer()
	defer ps.ShutdownNow()

	err := ps.Start()
	require.NoError(t, err, "unable to start plik server")

	err = ps.Start()
	require.Error(t, err, "should not be able to start plik server twice")
	require.Equal(t, "can't start a Plik server twice", err.Error(), "invalid error")

	err = ps.ShutdownNow()
	require.NoError(t, err, "unable to shutdown plik server")

	err = ps.ShutdownNow()
	require.Error(t, err, "should not be able to shutdown plik server twice")
	require.Equal(t, "can't shutdown a Plik server twice", err.Error(), "invalid error")

	err = ps.Start()
	require.Error(t, err, "should not be able to start a shutdown plik server")
	require.Equal(t, "can't start a shutdown Plik server", err.Error(), "invalid error")
}

func TestNewPlikServerNoHTTPSCertificates(t *testing.T) {
	ps := NewPlikServer(common.NewConfiguration())
	ps.config.ListenAddress = "127.0.0.1"
	ps.config.ListenPort = 44142
	ps.config.AutoClean(false)
	ps.config.MetadataBackend = "testing"
	ps.config.DataBackend = "testing"

	ps.config.SslEnabled = true

	err := ps.Start()
	require.Error(t, err, "unable to start plik server without ssl certificates")
}

func TestNewPlikServerWithCustomBackends(t *testing.T) {
	ps := newPlikServer()
	defer ps.ShutdownNow()

	ps.WithMetadataBackend(metadata_test.NewBackend())
	err := ps.initializeMetadataBackend()
	require.NoError(t, err, "invalid error")
	require.NotNil(t, ps.GetMetadataBackend(), "missing metadata backend")

	ps.WithDataBackend(data_test.NewBackend())
	err = ps.initializeDataBackend()
	require.NoError(t, err, "invalid error")
	require.NotNil(t, ps.GetDataBackend(), "missing data backend")

	ps.WithStreamBackend(data_test.NewBackend())
	err = ps.initializeStreamBackend()
	require.NoError(t, err, "invalid error")
	require.NotNil(t, ps.GetStreamBackend(), "missing stream backend")

}

func TestClean(t *testing.T) {
	ps := newPlikServer()
	defer ps.ShutdownNow()

	upload := common.NewUpload()
	upload.Create()
	upload.TTL = 1
	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()
	require.True(t, upload.IsExpired(), "upload should be expired")

	err := ps.metadataBackend.Upsert(ps.NewContext(), upload)
	require.NoError(t, err, "unable to save upload")

	ps.Clean()

	_, err = ps.metadataBackend.Get(ps.NewContext(), upload.ID)
	require.Error(t, err, "should be unable to get expired upload after clean")

	ps.metadataBackend.(*metadata_test.MetadataBackend).SetError(errors.New("error"))
	ps.Clean()
}

func TestAutoClean(t *testing.T) {
	ps := newPlikServer()
	defer ps.ShutdownNow()

	ps.cleaningRandomDelay = 1
	ps.cleaningMinOffset = 1
	ps.config.AutoClean(true)

	err := ps.Start()
	require.NoError(t, err, "unable to start plik server")

	upload := common.NewUpload()
	upload.Create()
	upload.TTL = 1
	upload.Creation = time.Now().Add(-10 * time.Minute).Unix()
	require.True(t, upload.IsExpired(), "upload should be expired")

	err = ps.metadataBackend.Upsert(ps.NewContext(), upload)
	require.NoError(t, err, "unable to save upload")

	time.Sleep(2 * time.Second)

	_, err = ps.metadataBackend.Get(ps.NewContext(), upload.ID)
	require.Error(t, err, "should be unable to get expired upload after clean")
}
