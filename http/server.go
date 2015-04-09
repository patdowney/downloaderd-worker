package http

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Config ...
type Config struct {
	ListenAddress string
}

// Server ...
type Server struct {
	ListenAddress   string
	Router          *mux.Router
	AccessLogWriter io.Writer
}

// NewServer ...
func NewServer(config *Config, accessLogWriter io.Writer) *Server {
	return &Server{
		ListenAddress:   config.ListenAddress,
		Router:          mux.NewRouter(),
		AccessLogWriter: accessLogWriter}
}

// AddResource ...
func (s *Server) AddResource(pathPrefix string, r Resource) {
	subrouter := s.Router.PathPrefix(pathPrefix).Subrouter()
	r.RegisterRoutes(subrouter)
}

func (s *Server) handler() http.Handler {
	return handlers.CombinedLoggingHandler(s.AccessLogWriter, s.Router)
}

// ListenAndServe ...
func (s *Server) ListenAndServe() error {
	http.Handle("/", s.handler())

	return http.ListenAndServe(s.ListenAddress, nil)
}
