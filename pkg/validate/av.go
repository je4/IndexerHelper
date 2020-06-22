package validate

import (
	"bytes"
	"context"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/memoriav/memobase-2020/services/histogram/pkg/service"
	"os/exec"
	"strings"
	"time"
)

type ValidateAV struct {
	log *logging.Logger
	ffmpeg  string
	timeout time.Duration
	wsl     bool
}

func NewValidateAV(ffmpeg string, timeout time.Duration, wsl bool, log *logging.Logger) (*ValidateAV, error) {
	h := &ValidateAV{
		ffmpeg:  ffmpeg,
		timeout: timeout,
		wsl:     wsl,
		log: log,
	}

	return h, nil
}

func (h *ValidateAV) Exec(file string, args ...interface{}) (interface{}, error) {
	colors := make(map[string]int64)

	if h.wsl {
		file = service.Path2Wsl(file)
	}

	cmdparam := []string{"-v", "warning", "-i", file, "-f", "null", "-"}
	cmdfile := h.ffmpeg
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
			return colors, emperror.Wrapf(err, "error executing (%s %s): %v - %v", cmdfile, cmdparam, outStr, errstr)
		}
	}

	//	data := out.String()
	errstr := strings.TrimSpace(errbuf.String())

	if isError {
		return struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "error",
			Message: errstr,
		}, nil
	}

	return struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "ok",
		Message: errstr,
	}, nil
}
