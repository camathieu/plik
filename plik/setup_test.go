package plik

import (
	"encoding/json"
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/data/file"
	"github.com/root-gg/plik/server/data/swift"
	data_test "github.com/root-gg/plik/server/data/testing"
	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/plik/server/server"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

//
// /!\ Backends ARE NOT automatically cleared between tests /!\
//
var metadataBackend *metadata.Backend
var dataBackend data.Backend

func TestMain(m *testing.M) {
	var err error

	// Setup cleaning
	code := 0
	cleanMetadata := func() {}
	cleanData := func() {}
	defer func() {
		cleanMetadata()
		cleanData()
		os.Exit(code)
	}()

	var testConfig *common.Configuration
	testConfigPath := os.Getenv("PLIKD_CONFIG")
	if testConfigPath != "" {
		fmt.Println("loading test config : " + testConfigPath)
		testConfig, err = common.LoadConfiguration(testConfigPath)
		if err != nil {
			fmt.Printf("Unable to load test configuration : %s\n", err)
			os.Exit(1)
		}
	} else {
		testConfig = common.NewConfiguration()
		testConfig.DataBackend = "testing"
		if os.Getenv("data_backend") != "" {
			testConfig.DataBackend = os.Getenv("data_backend")
			if os.Getenv("data_backend_config") != "" {
				var dataBackendConfig = make(map[string]interface{})
				err = json.Unmarshal([]byte(os.Getenv("data_backend_config")), &dataBackendConfig)
				if err != nil {
					fmt.Printf("Unable to deserialize data_backend_config : %s\n", err)
					os.Exit(1)
				}
			}
		}
	}

	// Setup data backend
	switch testConfig.DataBackend {
	case "file":
		dir, err := ioutil.TempDir("", "pliktest_file_")
		if err != nil {
			fmt.Printf("Unable to setup file data backend : %s\n", err)
			os.Exit(1)
		}

		cleanData = func() {
			err := os.RemoveAll(dir)
			if err != nil {
				fmt.Println(err)
			}
		}

		dataBackend = file.NewBackend(&file.Config{Directory: dir})
		fmt.Println("running tests with file data backend")
	case "swift":
		swiftConfig := swift.NewConfig(testConfig.DataBackendConfig)
		dataBackend = swift.NewBackend(swiftConfig)
		fmt.Println("running tests with swift data backend")
	case "testing":
		dataBackend = data_test.NewBackend()
	default:
		fmt.Printf("Invalid metadata backend : %s\n", testConfig.DataBackend)
		os.Exit(1)
	}

	// Run tests
	code = m.Run()
	os.Exit(code)
}

//
// /!\ Backends ARE NOT automatically cleared between tests /!\
//
func newPlikServerAndClient() (ps *server.PlikServer, pc *Client) {
	config := common.NewConfiguration()
	config.ListenAddress = "127.0.0.1"
	config.ListenPort = common.APIMockServerDefaultPort
	config.AutoClean(false)
	//config.Debug = true
	_ = config.Initialize()
	ps = server.NewPlikServer(config)

	metadataBackendConfig := &metadata.Config{Driver: "sqlite3", ConnectionString: "plik.test.db", EraseFirst: true}
	metadataBackend, err := metadata.NewBackend(metadataBackendConfig)
	if err != nil {
		panic(err)
	}
	ps.WithMetadataBackend(metadataBackend)

	ps.WithDataBackend(dataBackend)
	pc = NewClient(config.GetServerURL().String())
	return ps, pc
}

//
// /!\ Backends ARE NOT automatically cleared between tests /!\
//
func start(ps *server.PlikServer) (err error) {
	//err = common.CheckHTTPServer(ps.GetConfig().ListenPort)
	//if err == nil {
	//	return fmt.Errorf("plik server is already running")
	//}

	err = ps.Start()
	if err != nil {
		return err
	}

	err = common.CheckHTTPServer(ps.GetConfig().ListenPort)
	if err != nil {
		return err
	}

	return nil
}

//
// /!\ Backends ARE NOT automatically cleared between tests /!\
//
func shutdown(ps *server.PlikServer) {
	err := ps.ShutdownNow()
	if err != nil {
		panic("unable to shutdown server " + err.Error())
	}
	//err = common.CheckHTTPServer(ps.GetConfig().ListenPort)
	//if err == nil {
	//	panic("still able to join plik server after shutdown")
	//}
}

type LockedReader struct {
	lock   chan struct{}
	reader io.Reader
}

func NewLockedReader(reader io.Reader) (lr *LockedReader) {
	lr = new(LockedReader)
	lr.lock = make(chan struct{})
	lr.reader = reader
	return lr
}

func (lr *LockedReader) Read(p []byte) (n int, err error) {
	<-lr.lock
	return lr.reader.Read(p)
}

func (lr *LockedReader) Unleash() {
	close(lr.lock)
}

func NewSlowReaderRandom(reader io.Reader) (lr *LockedReader) {
	timeout := time.Duration(rand.Intn(1000)) * time.Millisecond
	return NewSlowReader(reader, timeout)
}

func NewSlowReader(reader io.Reader, timeout time.Duration) (lr *LockedReader) {
	lr = NewLockedReader(reader)
	go func() {
		<-time.After(timeout)
		lr.Unleash()
	}()
	return lr
}
