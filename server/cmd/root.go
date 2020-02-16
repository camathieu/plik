package cmd

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/plik/server/server"
	"github.com/root-gg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var configPath string
var config *common.Configuration
var port int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "plikd",
	Short: "Plik server",
	Long: `Plik server`,
	Version: common.GetBuildInfo().String(),
	Run: startPlikServer,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (default is /etc/plikd.cfg)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().IntVar(&port, "port", 0, "Overrides plik listen port")
}

// initConfig load configuration
func initConfig() {
	var err error

	if configPath != "" {
		// Use config file from the flag.
		info, err := os.Stat(configPath)
		if err != nil {
			fmt.Printf("Unable to read config file %s : %s\n", configPath, err)
			os.Exit(1)
		}
		if info.IsDir() {
			fmt.Printf("Unable to read config file %s\n", configPath)
			os.Exit(1)
		}
	} else {
		configPath = os.Getenv("PLIKD_CONFIG")
		if configPath != "" {
			// Use config file from env.
			info, err := os.Stat(configPath)
			if err != nil {
				fmt.Printf("Unable to read config file %s : %s\n", configPath, err)
				os.Exit(1)
			}
			if info.IsDir() {
				fmt.Printf("Unable to read config file %s\n", configPath)
				os.Exit(1)
			}
		} else {
			// Use config file from default locations.
			info, err := os.Stat("plikd.cfg")
			if err == nil && !info.IsDir() {
				configPath = "plikd.cfg"
			} else {
				info, err := os.Stat("/etc/plikd.cfg")
				if err == nil && !info.IsDir() {
					configPath = "/etc/plikd.cfg"
				}
			}
		}
	}

	if configPath != "" {
		config, err = common.LoadConfiguration(configPath)
		if err != nil {
			fmt.Printf("Unable to read config file : %s\n", err)
		}
	} else {
		config = common.NewConfiguration()
	}
}

// Initialize metadata backend
var initializeMetadataBackendOnce sync.Once
var metadataBackend *metadata.Backend
func initializeMetadataBackend() {
	var err error
	initializeMetadataBackendOnce.Do(func(){
		config := metadata.NewConfig(config.MetadataBackendConfig)
		metadataBackend, err = metadata.NewBackend(config)
		if err != nil {
			fmt.Printf("unable to initialize metadata backend : %s\n", err)
			os.Exit(1)
		}
	})
}

func startPlikServer(cmd *cobra.Command, args []string) {
	// Overrides port if provided in command line
	if port != 0 {
		config.ListenPort = port
	}

	if config.Debug {
		utils.Dump(config)
	}

	plik := server.NewPlikServer(config)

	err := plik.Start()
	if err != nil {
		fmt.Printf("unable to start Plik server : %s\n", err)
		os.Exit(1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_ = plik.Shutdown(time.Minute)
		os.Exit(0)
	}()

	select {}
}