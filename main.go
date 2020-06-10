package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/goph/emperror"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	configFile := flag.String("cfg", "./histogram.toml", "config file location")
	imgFile := flag.String("img", "", "config file location")
	flag.Parse()

	var exPath = ""
	// if configfile not found try path of executable as prefix
	if !FileExists(*configFile) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath = filepath.Dir(ex)
		if FileExists(filepath.Join(exPath, *configFile)) {
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
	log, lf := CreateLogger("indexer", config.Logfile, config.Loglevel)
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

	// if necessary create a colormap image
	if !FileExists(config.ImageMagick.Remap) {
		width := len(config.Colormap)
		height := 1
		upLeft := image.Point{0, 0}
		lowRight := image.Point{width, height}
		img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
		x := 0
		for name, value := range config.Colormap {
			color, err := ParseHexColor(value)
			if err != nil {
				log.Panicf("cannot parse color %s: %s", name, value)
				return
			}
			img.Set(x, 0, color)
			x++
		}
		f, err := os.Create(config.ImageMagick.Remap)
		if err != nil {
			log.Panicf("cannot create file %s: %v", config.ImageMagick.Remap, err)
			return
		}
		png.Encode(f, img)
		f.Close()
	}
	histogram, err := getHistogram(
		config.ImageMagick.Convert,
		config.ImageMagick.Resize,
		config.ImageMagick.Remap,
		config.ImageMagick.Colors,
		*imgFile,
		config.ImageMagick.Timeout.Duration,
		config.ImageMagick.Wsl)
	if err != nil {
		log.Panicf("error getting histogram: %v", err)
		return
	}
	result := make(map[string]int64)
	for col, weight := range histogram {
		ok := false
		for name, hex := range config.Colormap {
			if col == hex {
				ok = true
				result[name] = weight
			}
		}
		if !ok {
			log.Errorf("color %s not in colormap", col)
		}
	}

	json, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		log.Errorf("cannot marshal result: %v", err)
	}
	fmt.Println(string(json))
}

func getHistogram(convert, resize, remap string, colors int, file string, timeout time.Duration, wsl bool) (map[string]int64, error) {
	result := make(map[string]int64)

	cmdparam := []string{file, "-resize", resize, "-dither", "Riemersma", "-colors", fmt.Sprintf("%d", colors), "+dither", "-remap", remap, "-format", `%c`, "histogram:info:"}
	cmdfile := convert
	if wsl {
		cmdparam = append([]string{cmdfile}, cmdparam...)
		cmdfile = "wsl"
	}

	var out bytes.Buffer
	out.Grow(1024 * 1024) // 1MB size
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdfile, cmdparam...)
	//	cmd.Stdin = dataOut
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return result, emperror.Wrapf(err, "error executing (%s %s): %v - %v", cmdfile, cmdparam, out.String(), err)
	}

	data := out.String()
	//      44391: (  0,  0,  0,  0) #00000000 none
	//       182: (  0,  0,  0,  2) #00000002 srgba(0,0,0,0.00784314)
	scanner := bufio.NewScanner(strings.NewReader(data))
	r := regexp.MustCompile(` ([0-9]+):.+(#[0-9A-F]{6})[0-9A-F]{2} `)
	for scanner.Scan() {
		line := scanner.Text()
		matches := r.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		col := matches[2]
		countstr := matches[1]
		if _, ok := result[col]; !ok {
			result[col] = 0
		}
		count, err := strconv.ParseInt(countstr, 10, 64)
		if err != nil {
			return result, emperror.Wrapf(err, "cannot parse number %s in line %s", countstr, line)
		}
		result[col] += count
	}

	return result, nil
}
