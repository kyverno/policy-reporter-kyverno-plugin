package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting"
)

var funcMap = template.FuncMap{
	"add": func(i, j int) int {
		return i + j
	},
}

// PolicyHandler for the PolicyReport REST API
func PolicyReportingHandler(s reporting.PolicyReportGenerator, basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := s.PerPolicyData(req.Context(), reporting.Filter{
			Namespaces:   req.URL.Query()["namespaces"],
			Policies:     req.URL.Query()["policies"],
			ClusterScope: req.URL.Query().Get("clusterScope") != "0",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("policy-report-details.html").Funcs(funcMap).ParseFiles(path.Join(basePath, "policy-report-details.html"), path.Join(basePath, "mui.css"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// NamespaceReportingHandler for the NamespaceReport REST API
func NamespaceReportingHandler(s reporting.PolicyReportGenerator, basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := s.PerNamespaceData(req.Context(), reporting.Filter{
			Namespaces:   req.URL.Query()["namespaces"],
			Policies:     req.URL.Query()["policies"],
			ClusterScope: req.URL.Query().Get("clusterScope") != "0",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("namespace-report-details.html").Funcs(funcMap).ParseFiles(path.Join(basePath, "namespace-report-details.html"), path.Join(basePath, "mui.css"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

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
