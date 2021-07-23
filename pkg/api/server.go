package api

import (
	"fmt"
	"net/http"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// Server for the optional HTTP REST API
type Server interface {
	// Start the HTTP REST API
	Start() error
}

type httpServer struct {
	port  int
	mux   *http.ServeMux
	store *kyverno.PolicyStore
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/policies", Gzip(PolicyHandler(s.store)))
	s.mux.HandleFunc("/healthz", HealthzHandler(s.store))
	s.mux.HandleFunc("/ready", ReadyHandler())
}

func (s *httpServer) Start() error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	return server.ListenAndServe()
}

// NewServer constructor for a new API Server
func NewServer(pStore *kyverno.PolicyStore, port int) Server {
	s := &httpServer{
		port:  port,
		store: pStore,
		mux:   http.NewServeMux(),
	}

	s.registerHandler()

	return s
}
