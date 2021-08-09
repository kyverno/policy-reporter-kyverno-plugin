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
	Healthy()
	Unhealthy()
}

type httpServer struct {
	port        int
	mux         *http.ServeMux
	store       *kyverno.PolicyStore
	healthy     bool
	healthyChan chan bool
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/policies", Gzip(PolicyHandler(s.store)))
	s.mux.HandleFunc("/healthz", HealthzHandler(&s.healthy))
	s.mux.HandleFunc("/ready", ReadyHandler())
}

func (s *httpServer) Start() error {
	go func() {
		for healthy := range s.healthyChan {
			s.healthy = healthy
		}
	}()

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	return server.ListenAndServe()
}

func (s *httpServer) Healthy() {
	s.healthyChan <- true
}

func (s *httpServer) Unhealthy() {
	s.healthyChan <- false
}

// NewServer constructor for a new API Server
func NewServer(pStore *kyverno.PolicyStore, port int) Server {
	s := &httpServer{
		port:        port,
		store:       pStore,
		mux:         http.NewServeMux(),
		healthyChan: make(chan bool, 2),
	}

	s.registerHandler()

	return s
}
