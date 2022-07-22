package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// PolicyHandler for the Policy REST API
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

// VerifyImageRulesHandler for the ImageVerify Policy REST API
func VerifyImageRulesHandler(s *kyverno.PolicyStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		policies := s.List()
		if len(policies) == 0 {
			fmt.Fprint(w, "[]")

			return
		}

		verifyRules := make([]*VerifyImage, 0)
		images := map[string]bool{}

		for _, policy := range policies {
			for _, rule := range policy.Rules {
				if rule.VerifyImages == nil || len(rule.VerifyImages) == 0 {
					continue
				}

				for _, verify := range rule.VerifyImages {
					if _, ok := images[verify.Image]; ok {
						continue
					}

					verifyRules = append(verifyRules, &VerifyImage{
						Policy:       &Policy{Name: policy.Name, Namespace: policy.Namespace, UID: policy.UID},
						Rule:         rule.Name,
						Repository:   verify.Repository,
						Image:        verify.Image,
						Key:          verify.Key,
						Attestations: verify.Attestations,
					})

					images[verify.Image] = true
				}
			}
		}

		if err := json.NewEncoder(w).Encode(verifyRules); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{ "message": "%s" }`, err.Error())
		}
	}
}

// HealthzHandler for the Liveness REST API
func HealthzHandler(synced func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !synced() {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			log.Println("[WARNING] - Healthz Check: No kyverno policy crds are found")

			fmt.Fprint(w, `{ "error": "No policy CRDs found" }`)

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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
