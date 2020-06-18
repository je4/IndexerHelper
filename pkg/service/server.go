package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goph/emperror"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
)

type Server struct {
	srv       *http.Server
	host      string
	port      string
	log       *logging.Logger
	accesslog io.Writer
	services  map[string]Service
	wsl       bool
}

func NewServer(
	addr string,
	log *logging.Logger,
	accesslog io.Writer,
	srvs map[string]Service,
	wsl bool,
) (*Server, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		//log.Panicf("cannot split address %s: %v", addr, err)
		return nil, emperror.Wrapf(err, "cannot split address %s", addr)
	}

	srv := &Server{
		srv:       nil,
		host:      host,
		port:      port,
		log:       log,
		accesslog: accesslog,
		services:  srvs,
		wsl:       wsl,
	}
	return srv, nil
}

func (s *Server) ListenAndServe(cert, key string) error {
	router := mux.NewRouter()

	srvRegexp := regexp.MustCompile(`^/([^/]+)/(.+)$`)
	router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		matches := srvRegexp.FindSubmatch([]byte(r.URL.Path))
		if len(matches) == 0 {
			return false
		}
		rm.Vars = map[string]string{}
		prefix := string(matches[1])
		rm.Vars["prefix"] = prefix
		rm.Vars["path"] = string(matches[2])
		// check whether prefix is known
		for key, _ := range s.services {
			if key == prefix {
				return true
			}
		}
		return false
	}).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(request)
		prefix, ok := vars["prefix"]
		if !ok {
			s.DoPanicf(writer, http.StatusNotFound, "invalid url %s. no path given", request.URL.String())
			return
		}
		serv, ok := s.services[prefix]
		if !ok {
			s.DoPanicf(writer, http.StatusNotFound, "invalid url %s. wrong prefix %s", request.URL.String(), prefix)
			return
		}
		path, ok := vars["path"]
		if !ok {
			s.DoPanicf(writer, http.StatusNotFound, "invalid url %s. no path given", request.URL.String())
			return
		}
		if runtime.GOOS != "windows" || (runtime.GOOS == "windows" && s.wsl) {
			path = `/` + path
		}
		p, err := url.QueryUnescape(path)
		if err != nil {
			s.DoPanicf(writer, http.StatusInternalServerError, "cannot unescape path %s: %v", path, err)
			return
		}
		result, err := serv.Exec(p)
		if err != nil {
			s.DoPanicf(writer, http.StatusInternalServerError, "cannot create %s of %s: %v", prefix, path, err)
			return
		}
		jw := json.NewEncoder(writer)
		jw.Encode(result)
	}).Methods("GET")

	loggedRouter := handlers.LoggingHandler(s.accesslog, router)
	addr := net.JoinHostPort(s.host, s.port)
	s.srv = &http.Server{
		Handler: loggedRouter,
		Addr:    addr,
	}
	if cert != "" && key != "" {
		s.log.Infof("starting HTTPS histogram server at https://%v", addr)
		return s.srv.ListenAndServeTLS(cert, key)
	} else {
		s.log.Infof("starting HTTP histogram server at http://%v", addr)
		return s.srv.ListenAndServe()
	}
}

func (s *Server) DoPanicf(writer http.ResponseWriter, status int, message string, a ...interface{}) (err error) {
	msg := fmt.Sprintf(message, a...)
	s.DoPanic(writer, status, msg)
	return
}

func (s *Server) DoPanic(writer http.ResponseWriter, status int, message string) (err error) {
	type errData struct {
		Status     int
		StatusText string
		Message    string
	}
	s.log.Error(message)
	data := errData{
		Status:     status,
		StatusText: http.StatusText(status),
		Message:    message,
	}
	writer.WriteHeader(status)
	// if there'ms no error Template, there's no help...
	jw := json.NewEncoder(writer)
	jw.Encode(data)
	return
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
