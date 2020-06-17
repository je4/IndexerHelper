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
	Wsl      bool
	Timeout  duration
	Remap    string
	Colors   int
	Resize   string
}

type Config struct {
	Logfile         string
	Loglevel        string
	AccessLog       string
	CertPEM         string
	KeyPEM          string
	Addr            string
	HistogramPrefix string
	JwtKey          string
	JwtAlg          []string
	ImageMagick     ConfigImageMagick
	Colormap        map[string]string
}

func LoadConfig(filepath string) Config {
	var conf Config
	_, err := toml.DecodeFile(filepath, &conf)
	if err != nil {
		log.Fatalln("Error on loading config: ", err)
	}
	return conf
}
