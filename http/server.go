package http

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type HTTPConfig struct {
	ListenAddress string
}

type Server struct {
	ListenAddress   string
	Router          *mux.Router
	AccessLogWriter io.Writer
}

func NewServer(config *HTTPConfig, accessLogWriter io.Writer) *Server {
	return &Server{
		ListenAddress:   config.ListenAddress,
		Router:          mux.NewRouter(),
		AccessLogWriter: accessLogWriter}
}

func (s *Server) AddResource(pathPrefix string, r Resource) {
	subrouter := s.Router.PathPrefix(pathPrefix).Subrouter()
	r.RegisterRoutes(subrouter)
}

func (s *Server) handler() http.Handler {
	return handlers.CombinedLoggingHandler(s.AccessLogWriter, s.Router)
}

func (s *Server) ListenAndServe() error {
	http.Handle("/", s.handler())

	return http.ListenAndServe(s.ListenAddress, nil)
}
