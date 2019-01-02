package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/google/logger"
	"github.com/ksaylor11/GoDVR/config"
	"github.com/ksaylor11/go-apache-log-parser"
	"net/http"
	"os"
	"time"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")
var logPathInput = flag.String("output", "audit.log", "output location")
var confFileInput = flag.String("config", "config.toml", "configuration file")

type LineList struct {
	Items []apachelogparser.Line
}

func main() {

	var (
		conf config.Config
	)
	configFile := *confFileInput
	logPath := *logPathInput

	// setting up our logging
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

	// parse our apache log
	lines, err := apachelogparser.Parse(conf.Filename)
	if err != nil {
		logger.Fatal(err)
	}

	// construct a list to work with instead of reading the file
	l := LineList{}

	StartTime,err := time.Parse("2006-01-02 15:04:05", conf.StartTime)
	EndTime,err := time.Parse("2006-01-02 15:04:05", conf.EndTime)
	for _, line := range lines {
		if line.Time.After(StartTime) && line.Time.Before(EndTime){
			l.Items = append(l.Items, line)
		}
	}

	fmt.Printf("number of lines: %d\n", len(l.Items))

	PreviousTime := StartTime
	for _, line := range l.Items {
		// get current time
		if line.Method == "GET" {
			if line.Time.After(PreviousTime) {
				time.Sleep(line.Time.Sub(PreviousTime))
				result, err := replay(line.URL, conf.Host, conf.Protocol)
				if err != nil {
					logger.Fatal(err)
				}
				fmt.Printf("url: %s \t status: %d\n", line.URL, result)
				PreviousTime = line.Time
			} else{
				result, err := replay(line.URL, conf.Host, conf.Protocol)
				if err != nil {
					logger.Fatal(err)
				}
				fmt.Printf("url: %s \t status: %d\n", line.URL, result)
				PreviousTime = line.Time
			}
		} else {
			logger.Infof("Skipping %s requests\n", line.Method)
		}
	}
}

// replay function
func replay(URL string, host string, protocol string) (code int, err error) {
	// assemble the url
	fullurl := fmt.Sprintf("%s://%s%s", protocol, host, URL)
	logger.Infof("fetching: %s\n", fullurl)

	// make the request
	resp, err := http.Get(fullurl)
	logger.Infof("result code %d\n", resp.StatusCode)

	return resp.StatusCode, err
}
