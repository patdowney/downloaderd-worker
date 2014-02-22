package http

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

type HTTPConfig struct {
	ListenAddress string
}

type Server struct {
	ListenAddress   string
	Router          *mux.Router
	AccessLogWriter io.Writer
}

func NewServer(config *HTTPConfig) *Server {
	return &Server{
		ListenAddress:   config.ListenAddress,
		Router:          mux.NewRouter(),
		AccessLogWriter: os.Stdout}
}

func (s *Server) AddResource(pathPrefix string, r Resource) {
	subrouter := s.Router.PathPrefix(pathPrefix).Subrouter()
	r.RegisterRoutes(subrouter)
}

func (s *Server) GetHandler() http.Handler {
	return handlers.CombinedLoggingHandler(s.AccessLogWriter, s.Router)
}

func (s *Server) ListenAndServe() error {
	http.Handle("/", s.GetHandler())

	return http.ListenAndServe(s.ListenAddress, nil)
}
