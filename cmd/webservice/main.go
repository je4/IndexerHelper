package main

import (
	"context"
	"flag"
	"fmt"
	exif2 "gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/exif"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/histogram"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/service"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/validate"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	configFile := flag.String("cfg", "./histogram.toml", "config file location")
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

	var accesslog io.Writer
	if config.AccessLog == "" {
		accesslog = os.Stdout
	} else {
		f, err := os.OpenFile(config.AccessLog, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Panicf("cannot open file %s: %v", config.AccessLog, err)
			return
		}
		defer f.Close()
		accesslog = f
	}
	fmt.Sprintf("%v", accesslog)

	imgval, err := validate.NewValidateImage(config.ImageMagick.Identify, config.ValidateImage.Timeout.Duration, config.Wsl, log)
	if err != nil {
		log.Fatalf("cannot initialize ValidateImage: %v", err)
		return
	}

	avval, err := validate.NewValidateAV(config.FFMPEG.FFMPEG, config.ValidateAV.Timeout.Duration, config.Wsl, log)
	if err != nil {
		log.Fatalf("cannot initialize ValidateAV: %v", err)
		return
	}

	exif, err := exif2.NewExif(
		config.Exiftool.Exiftool,
		config.Exif.Params,
		config.Exif.Timeout.Duration,
		config.Wsl,
		log)
	if err != nil {
		log.Fatalf("cannot initialize Exif: %v", err)
		return
	}

	hist, err := histogram.NewHistogram(
		config.ImageMagick.Convert,
		config.Histogram.Resize,
		config.Histogram.Remap,
		config.Histogram.Colormap,
		config.Histogram.Colors,
		config.Histogram.Timeout.Duration,
		config.Wsl)
	if err != nil {
		log.Fatalf("cannot initialize histogram: %v", err)
		return
	}

	srv, err := service.NewServer(
		config.Addr,
		log,
		accesslog,
		map[string]service.Service{
			config.Histogram.Prefix:     hist,
			config.ValidateImage.Prefix: imgval,
			config.ValidateAV.Prefix:    avval,
			config.Exif.Prefix:          exif,
		},
		config.Wsl,
	)
	if err != nil {
		log.Fatalf("cannot initialize server: %v", err)
		return
	}
	go func() {
		if err := srv.ListenAndServe(config.CertPEM, config.KeyPEM); err != nil {
			log.Errorf("server died: %v", err)
		}
	}()

	end := make(chan bool, 1)

	// process waiting for interrupt signal (TERM or KILL)
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		log.Infof("shutdown requested")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		end <- true
	}()

	<-end
	log.Info("server stopped")
}
