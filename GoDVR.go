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

	// adding rate limiting
	limit := conf.Limit
	if limit == 0 {
		for _, line := range lines {
			timeReplay(line, conf)
		}
	} else {
		for _, line := range lines[0:limit] {
			timeReplay(line, conf)
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

// replays within time frame
func timeReplay(line apachelogparser.Line, conf config.Config){
	if line.Method == "GET" {
		// get current time
		CurrentTime := time.Now()
		ModifiedTime := line.Time.Add(time.Hour * 24)
		if ModifiedTime.After(CurrentTime) {
			time.Sleep(ModifiedTime.Sub(CurrentTime))
			result, err := replay(line.URL, conf.Host, conf.Protocol)
			if err != nil {
				logger.Fatal(err)
			}
			fmt.Printf("url: %s \t status: %d\n", line.URL, result)
		} else {
			logger.Infof("skipping prior requests. time:%s\n", line.Time.Format("2006-01-02 15:04:05"))
		}
	} else {
		logger.Infof("Skipping %s requests\n", line.Method)
	}
}
