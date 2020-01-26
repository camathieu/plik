package main

import (
	"flag"
	"fmt"
	"github.com/root-gg/utils"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var configFile = flag.String("config", "plikd.cfg", "Configuration file (default: plikd.cfg")
	var version = flag.Bool("version", false, "Show version of plikd")
	var port = flag.Int("port", 0, "Overrides plik listen port")
	flag.Parse()
	if *version {
		fmt.Printf("Plik server %s\n", common.GetBuildInfo())
		os.Exit(0)
	}

	config, err := common.LoadConfiguration(*configFile)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	// Overrides port if provided in command line
	if *port != 0 {
		config.ListenPort = *port
	}

	if config.LogLevel == "DEBUG" {
		utils.Dump(config)
	}

	plik := server.NewPlikServer(config)

	err = plik.Start()
	if err != nil {
		log.Fatal(err.Error())
		return
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
