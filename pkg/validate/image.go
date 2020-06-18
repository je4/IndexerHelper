package validate

import (
	"bytes"
	"context"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"os/exec"
	"strings"
	"time"
)

type ValidateImage struct {
	identify string
	timeout  time.Duration
	wsl      bool
	log *logging.Logger
}

func NewValidateImage(identify string, timeout time.Duration, wsl bool, log *logging.Logger) (*ValidateImage, error) {
	h := &ValidateImage{
		identify: identify,
		timeout:  timeout,
		wsl:      wsl,
		log: log,
	}

	return h, nil
}

func (h *ValidateImage) Exec(file string, args ...interface{}) (interface{}, error) {
	colors := make(map[string]int64)

	cmdparam := []string{"-verbose", file}
	cmdfile := h.identify
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

	if err := cmd.Run(); err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if ok && exiterr.ExitCode() == 1 {
			outStr := out.String()
			errstr := strings.TrimSpace(errbuf.String())
			return colors,emperror.Wrapf(err, "error executing (%s %s): %v - %v", cmdfile, cmdparam, outStr, errstr)
		} else {
			outStr := out.String()
			errstr := strings.TrimSpace(errbuf.String())
			return colors, emperror.Wrapf(err, "error executing (%s %s): %v - %v", cmdfile, cmdparam, outStr, errstr)
		}
	}

	//	data := out.String()
	errstr := strings.TrimSpace(errbuf.String())

	if len(errstr) > 0 {
		return struct{
			Status string `json:"status"`
			Message string `json:"message"`
		}{
			Status: "error",
			Message: errstr,
			}, nil
	}

	return struct{
		Status string `json:"status"`
		Message string `json:"message"`
	}{
		Status: "ok",
		Message: out.String(),
	}, nil
}
