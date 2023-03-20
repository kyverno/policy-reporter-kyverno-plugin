package api

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting"
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
	mux     *http.ServeMux
	store   *kyverno.PolicyStore
	reports reporting.PolicyReportGenerator
	http    http.Server
	synced  func() bool
}

func (s *httpServer) registerHandler() {
	s.mux.HandleFunc("/healthz", HealthzHandler(s.synced))
	s.mux.HandleFunc("/ready", ReadyHandler())
}

func (s *httpServer) RegisterMetrics() {
	s.mux.Handle("/metrics", promhttp.Handler())
}

func (s *httpServer) RegisterREST() {
	s.mux.HandleFunc("/policies", Gzip(PolicyHandler(s.store)))
	s.mux.HandleFunc("/verify-image-rules", Gzip(VerifyImageRulesHandler(s.store)))
	s.mux.HandleFunc("/namespace-details-reporting", Gzip(NamespaceReportingHandler(s.reports, path.Join("templates", "reporting"))))
	s.mux.HandleFunc("/policy-details-reporting", Gzip(PolicyReportingHandler(s.reports, path.Join("templates", "reporting"))))
}

func (s *httpServer) Start() error {
	return s.http.ListenAndServe()
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// NewServer constructor for a new API Server
func NewServer(pStore *kyverno.PolicyStore, reports reporting.PolicyReportGenerator, port int, synced func() bool, logger *zap.Logger) Server {
	mux := http.NewServeMux()

	s := &httpServer{
		store:   pStore,
		reports: reports,
		mux:     mux,
		synced:  synced,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: NewLoggerMiddleware(logger, mux),
		},
	}

	s.registerHandler()

	return s
}

func NewLoggerMiddleware(logger *zap.Logger, mux http.Handler) http.Handler {
	if logger == nil {
		return mux
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := []zapcore.Field{
			zap.String("proto", r.Proto),
			zap.String("user-agent", r.Header.Get("User-Agent")),
			zap.String("path", r.URL.Path),
		}

		if query := r.URL.RawQuery; query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if ref := r.Header.Get("Referer"); ref != "" {
			fields = append(fields, zap.String("referer", ref))
		}
		if scheme := r.URL.Scheme; scheme != "" {
			fields = append(fields, zap.String("scheme", scheme))
		}

		logger.Debug("Serve", fields...)

		mux.ServeHTTP(w, r)
	})
}
