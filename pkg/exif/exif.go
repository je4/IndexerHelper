package exif

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/service"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Exif struct {
	log      *logging.Logger
	exiftool string
	timeout  time.Duration
	wsl      bool
	params   []string
}

var paramregexp = regexp.MustCompile(`([a-zA-Z0-9]+:([^ ']+|'[^']+'))|([^ ']+)|'([^']+)'`)

func NewExif(exiftool string, params string, timeout time.Duration, wsl bool, log *logging.Logger) (*Exif, error) {
	h := &Exif{
		exiftool:  exiftool,
		timeout: timeout,
		wsl:     wsl,
		log:     log,
	}
	h.params = paramregexp.FindAllString(params, -1)
	if h.params == nil {
		h.params = []string{}
	}

	return h, nil
}

func (h *Exif) Exec(file string, args ...interface{}) (interface{}, error) {
	if h.wsl {
		file = service.Path2Wsl(file)
	}

	cmdparam := []string{}
	for _, p := range h.params {
		if p == "[[PATH]]" {
			p = file
		}
		cmdparam = append(cmdparam, p)
	}

	cmdfile := h.exiftool
	if h.wsl {
		cmdparam = append([]string{cmdfile}, cmdparam...)
		cmdfile = "wsl"
	}

	var out bytes.Buffer
	out.Grow(1024 * 1024) // 1MB size
	var errbuf bytes.Buffer
	errbuf.Grow(1024 * 1024) // 1MB size
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdfile, cmdparam...)
	//	cmd.Stdin = dataOut
	cmd.Stdout = &out
	cmd.Stderr = &errbuf

	h.log.Infof("executing %v %v", cmdfile, cmdparam)


	isError := false
	if err := cmd.Run(); err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if ok && exiterr.ExitCode() == 1 {
			isError = true
		} else {
			outStr := out.String()
			errstr := strings.TrimSpace(errbuf.String())
			return nil, emperror.Wrapf(err, "error executing (%s %s): %v - %v", cmdfile, cmdparam, outStr, errstr)
		}
	}
	outStr := out.String()
	errstr := strings.TrimSpace(errbuf.String())

	//	data := out.String()

	if isError {
		return struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "error",
			Message: errstr,
		}, nil
	}

	var result interface{}

	if err := json.Unmarshal([]byte(outStr), &result); err != nil {
		return nil, emperror.Wrapf(err, "cannot unmarshal data - %s", outStr)
	}

	ilist, ok := result.([]interface{})
	if ok {
		if len(ilist) != 1 {
			return nil, fmt.Errorf("invalid number of objects in json %v - %s", len(ilist), outStr)
		}
		result = ilist[0]
	}

	return result, nil
}
