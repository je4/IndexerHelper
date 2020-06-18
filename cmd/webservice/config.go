package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

type duration struct {
	Duration time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type ConfigImageMagick struct {
	Identify string
	Convert  string
}

type ConfigFFMPEG struct {
	FFMPEG  string
	FFPROBE string
}

type ConfigHistogram struct {
	Prefix   string
	Colormap map[string]string
	Timeout  duration
	Remap    string
	Colors   int
	Resize   string
}

type ConfigValidateImage struct {
	Prefix  string
	Timeout duration
}

type ConfigValidateAV struct {
	Prefix  string
	Timeout duration
}

type Config struct {
	Logfile       string
	Loglevel      string
	AccessLog     string
	CertPEM       string
	KeyPEM        string
	Addr          string
	JwtKey        string
	JwtAlg        []string
	ImageMagick   ConfigImageMagick
	FFMPEG        ConfigFFMPEG
	Histogram     ConfigHistogram
	ValidateImage ConfigValidateImage
	ValidateAV    ConfigValidateAV
	Wsl           bool
}

func LoadConfig(filepath string) Config {
	var conf Config
	_, err := toml.DecodeFile(filepath, &conf)
	if err != nil {
		log.Fatalln("Error on loading config: ", err)
	}
	return conf
}
