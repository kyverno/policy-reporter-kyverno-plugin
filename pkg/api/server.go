package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server for the optional HTTP REST API
type Server interface {
	// Start the HTTP REST API
	Start() error
	// Shutdown the HTTP Sever
	Shutdown(ctx context.Context) error
	// RegisterREST add all REST API handler
	RegisterREST()
	// RegisterMetrics adds Metrics handler
	RegisterMetrics()
}

type httpServer struct {
	mux            *http.ServeMux
	store          *kyverno.PolicyStore
	http           http.Server
	foundResources map[string]bool
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.foundResources))
	s.mux.HandleFunc("/ready", ReadyHandler())
}

func (s *httpServer) RegisterMetrics() {
	s.mux.Handle("/metrics", promhttp.Handler())
}

func (s *httpServer) RegisterREST() {
	s.mux.HandleFunc("/policies", Gzip(PolicyHandler(s.store)))
	s.mux.HandleFunc("/verify-image-rules", Gzip(VerifyImageRulesHandler(s.store)))
}

func (s *httpServer) Start() error {
	return s.http.ListenAndServe()
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// NewServer constructor for a new API Server
func NewServer(pStore *kyverno.PolicyStore, port int, foundResources map[string]bool) Server {
	mux := http.NewServeMux()

	s := &httpServer{
		store:          pStore,
		mux:            mux,
		foundResources: foundResources,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	s.registerHandler()

	return s
}
