package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// PolicyHandler for the PolicyReport REST API
func PolicyHandler(s *kyverno.PolicyStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		policies := s.List()
		if len(policies) == 0 {
			fmt.Fprint(w, "[]")

			return
		}

		if err := json.NewEncoder(w).Encode(policies); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{ "message": "%s" }`, err.Error())
		}
	}
}

// HealthzHandler for the Liveness REST API
func HealthzHandler(healthy *bool) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if *healthy == false {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{ "error": "Service unhealthy" }`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}
}

// ReadyHandler for the Readiness REST API
func ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}
}
