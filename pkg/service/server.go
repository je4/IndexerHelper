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
)

type Server struct {
	srv       *http.Server
	host      string
	port      string
	log       *logging.Logger
	accesslog io.Writer
	services  map[string]Service
}

func NewServer(
	addr string,
	log *logging.Logger,
	accesslog io.Writer,
	srvs map[string]Service,
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
	}
	return srv, nil
}

func (s *Server) ListenAndServe(cert, key string) error {
	router := mux.NewRouter()

	for prefix, serv := range s.services {

		router.HandleFunc(fmt.Sprintf("/%s/[path]", prefix), func(writer http.ResponseWriter, request *http.Request) {
			vars := mux.Vars(request)
			path, ok := vars["path"]
			if !ok {
				s.DoPanicf(writer, http.StatusNotFound, "invalid url %s. no path given", request.URL.String())
				return
			}
			result, err := serv.Exec(path)
			if err != nil {
				s.DoPanicf(writer, http.StatusInternalServerError, "cannot create histogram of %s", path)
				return
			}
			jw := json.NewEncoder(writer)
			jw.Encode(result)
		}).Methods("GET")
	}

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
