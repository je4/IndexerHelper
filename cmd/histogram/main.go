package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/histogram"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/service"
	"log"
	"os"
	"path/filepath"
)

func main() {
	configFile := flag.String("cfg", "./histogram.toml", "config file location")
	imgFile := flag.String("img", "", "config file location")
	flag.Parse()

	var exPath = ""
	// if configfile not found try path of executable as prefix
	if !service.FileExists(*configFile) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath = filepath.Dir(ex)
		if service.FileExists(filepath.Join(exPath, *configFile)) {
			*configFile = filepath.Join(exPath, *configFile)
		} else {
			log.Fatalf("cannot find configuration file: %v", *configFile)
			return
		}
	}
	// configfile should exists at this place
	var config Config
	config = LoadConfig(*configFile)

	// create logger instance
	log, lf := service.CreateLogger("indexer", config.Logfile, config.Loglevel)
	defer lf.Close()


	hist, err := histogram.NewHistogram(
		config.ImageMagick.Convert,
		config.Histogram.Resize,
		config.Histogram.Remap,
		config.Histogram.Colormap,
		config.Histogram.Colors,
		config.Histogram.Timeout.Duration,
		config.Wsl)

	// if necessary create a colormap image
	result, err := hist.Exec(*imgFile)
	if err != nil {
		log.Panicf("error getting histogram: %v", err)
		return
	}

	json, err := json.Marshal(result)
	if err != nil {
		log.Errorf("cannot marshal result: %v", err)
	}
	fmt.Println(string(json))
}

