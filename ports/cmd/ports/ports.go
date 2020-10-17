package main

import (
	"fmt"
	"github.com/kreyyser/transshipment/common/config"
	"github.com/spf13/pflag"
	"os"
)

const (
	success = iota
	initError
	runtimeError
)

func main() {
	code, msg := run()
	logExit(code, msg)
}

func run() (int, interface{}) {
	// Handle command line options
	flagset := pflag.NewFlagSet("ports", pflag.ExitOnError)
	configPath := flagset.StringP("config", "c", "", "Path to transhipment config yaml")
	_ = flagset.Parse(os.Args[1:])

	// Transhipment config file should exist in the file system
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		return initError, "no config file provided"
	}

	// Build the Config Manager
	cfmgr := config.NewManager(*configPath)

	// Load configuration file
	if err := cfmgr.Load(); err != nil {
		return initError, fmt.Sprintf("failed to load config file: %s", err)
	}

	// Build Ports Server
	server := NewServer(cfmgr)

	// Setup the configured Datastore
	if err := server.LoadStores(); err != nil {
		return initError, fmt.Sprintf("failed to load stores: %s", err)
	}

	// Setup the configured Ports Services
	if err := server.LoadServices(); err != nil {
		return initError, fmt.Sprintf("failed to load services: %s", err)
	}
	if err := server.Run(); err != nil {
		return runtimeError, fmt.Sprintf("failed to spin up service listeners. err: %v", err)
	}

	return success, "all good"
}

// logExit logs the termination message and returns the exit code
func logExit(code int, msg interface{}) {
	prefix := "msg"
	if code != 0 {
		prefix = "err"
	}
	fmt.Println(prefix, msg)
	os.Exit(code)
}
