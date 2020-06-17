package histogram

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/goph/emperror"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/service"
	"image"
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Histogram struct {
	convert, resize, remap string
	colors                 int
	colormap               map[string]string
	timeout                time.Duration
	wsl                    bool
}

func NewHistogram(convert, resize, remap string, colormap map[string]string, colors int, timeout time.Duration, wsl bool) (*Histogram, error) {
	h := &Histogram{
		convert:  convert,
		resize:   resize,
		remap:    remap,
		colors:   colors,
		colormap: colormap,
		timeout:  timeout,
		wsl:      wsl,
	}
	if !service.FileExists(remap) {
		width := len(colormap)
		height := 1
		upLeft := image.Point{0, 0}
		lowRight := image.Point{width, height}
		img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
		x := 0
		for name, value := range colormap {
			color, err := service.ParseHexColor(value)
			if err != nil {
				return nil, emperror.Wrapf(err, "cannot parse color %s: %s", name, value)
			}
			img.Set(x, 0, color)
			x++
		}
		f, err := os.Create(remap)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot create file %s", remap)
		}
		png.Encode(f, img)
		f.Close()
	}

	return h, nil
}
func (h *Histogram) GetName() string {
	return "histogram"
}
func (h *Histogram) Exec(file string, args ...interface{}) (interface{}, error) {
	colors := make(map[string]int64)

	cmdparam := []string{file, "-resize", h.resize, "-dither", "Riemersma", "-colors", fmt.Sprintf("%d", h.colors), "+dither", "-remap", h.remap, "-format", `%c`, "histogram:info:"}
	cmdfile := h.convert
	if h.wsl {
		cmdparam = append([]string{cmdfile}, cmdparam...)
		cmdfile = "wsl"
	}

	var out bytes.Buffer
	out.Grow(1024 * 1024) // 1MB size
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdfile, cmdparam...)
	//	cmd.Stdin = dataOut
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		outStr := out.String()
		return colors, emperror.Wrapf(err, "error executing (%s %s): %v", cmdfile, cmdparam, outStr)
	}

	data := out.String()
	//      44391: (  0,  0,  0,  0) #00000000 none
	//       182: (  0,  0,  0,  2) #00000002 srgba(0,0,0,0.00784314)
	scanner := bufio.NewScanner(strings.NewReader(data))
	r := regexp.MustCompile(` ([0-9]+):.+(#[0-9A-F]{6})[0-9A-F]{0,2} `)
	for scanner.Scan() {
		line := scanner.Text()
		matches := r.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		col := matches[2]
		countstr := matches[1]
		if _, ok := colors[col]; !ok {
			colors[col] = 0
		}
		count, err := strconv.ParseInt(countstr, 10, 64)
		if err != nil {
			return colors, emperror.Wrapf(err, "cannot parse number %s in line %s", countstr, line)
		}
		colors[col] += count
	}

	result := make(map[string]int64)
	for col, weight := range colors {
		ok := false
		for name, hex := range h.colormap {
			if col == hex {
				ok = true
				result[name] = weight
			}
		}
		if !ok {
			return nil, fmt.Errorf("color %s not in colormap", col)
		}
	}

	return result, nil
}
