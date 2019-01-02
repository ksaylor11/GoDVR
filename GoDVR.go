package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/google/logger"
	"github.com/ksaylor11/GoDVR/config"
	"github.com/ksaylor11/go-apache-log-parser"
	"os"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")
var logPathInput = flag.String("output", "chat_api.log", "output location")
var confFileInput = flag.String("config", "config.toml", "configuration file")

func main() {

	var (
		conf	config.Config
	)
	configFile := *confFileInput
	logPath := *logPathInput

	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()

	defer logger.Init("LoggerExample", *verbose, true, lf).Close()

	// setting up config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		logger.Fatalf("Error with config file: %v", err)
	}

	lines, err := apachelogparser.Parse(conf.Filename)
	if err != nil {
		logger.Fatal(err)
	}

	for _, line := range lines[0:1000] {
		if line.Method == "GET"{
			fmt.Printf("time: %s\n", line.Time)
			fmt.Printf("method: %s\n", line.Method)
			fmt.Printf("url: %s\n", line.URL)
		} else {
			logger.Infof("Skipping %s requests\n", line.Method)
		}

	}
}
