package api_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting"
)

type policyReportGeneratorStub struct {
	error error
}

func (g *policyReportGeneratorStub) PerPolicyData(ctx context.Context, filter reporting.Filter) ([]*reporting.Validation, error) {
	return []*reporting.Validation{
		{
			Name: "disallow-capabilities",
			Policy: &reporting.Policy{
				Title:       "Disallow Capabilities",
				Category:    "Pod Security Standards (Baseline)",
				Severity:    "medium",
				Description: "Adding capabilities beyond those listed in the policy must be disallowed.",
			},
			Groups: map[string]*reporting.Group{
				"kyverno": {
					Name: "kyverno",
					Rules: map[string]*reporting.Rule{
						"adding-capabilities": {
							Summary: &reporting.Summary{Pass: 1},
							Resources: []*reporting.Resource{{
								Name:       "kyverno",
								Kind:       "Deployment",
								APIVersion: "apps/v1",
								Status:     "pass",
							}},
						},
					},
				},
			},
		},
	}, g.error
}

func (g *policyReportGeneratorStub) PerNamespaceData(ctx context.Context, filter reporting.Filter) ([]*reporting.Validation, error) {
	return []*reporting.Validation{
		{
			Name: "kyverno",
			Groups: map[string]*reporting.Group{
				"disallow-capabilities": {
					Name: "disallow-capabilities",
					Policy: &reporting.Policy{
						Title:       "Disallow Capabilities",
						Category:    "Pod Security Standards (Baseline)",
						Severity:    "medium",
						Description: "Adding capabilities beyond those listed in the policy must be disallowed.",
					},
					Rules: map[string]*reporting.Rule{
						"adding-capabilities": {
							Summary: &reporting.Summary{Pass: 1},
							Resources: []*reporting.Resource{{
								Name:       "kyverno",
								Kind:       "Deployment",
								APIVersion: "apps/v1",
								Status:     "pass",
							}},
						},
					},
				},
			},
		},
	}, g.error
}

func Test_PolicyAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/policies", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.PolicyHandler(kyverno.NewPolicyStore()))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/policies", nil)
		if err != nil {
			t.Fatal(err)
		}

		result := &kyverno.Rule{
			ValidateMessage: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
			Name:            "autogen-check-for-requests-and-limits",
		}

		policy := kyverno.Policy{
			Kind:              "Policy",
			APIVersion:        "kyverno/v1",
			Name:              "require-ressources",
			Namespace:         "test",
			Rules:             []*kyverno.Rule{result},
			CreationTimestamp: time.Now(),
		}

		store := kyverno.NewPolicyStore()
		store.Add(policy)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.PolicyHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[{"kind":"Policy","apiVersion":"kyverno/v1","name":"require-ressources","namespace":"test","background":null,"rules":[{"message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","name":"autogen-check-for-requests-and-limits","type":""}]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_HealthzAPI(t *testing.T) {
	t.Run("Success Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.HealthzHandler(func() bool { return true }))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("Unavailable Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.HealthzHandler(func() bool { return false })

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
		}
	})
}

func Test_ReadyAPI(t *testing.T) {
	t.Run("Success Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/healthz", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.ReadyHandler())

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_VerifyImageRulesAPI(t *testing.T) {
	t.Run("Empty Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/verify-image-rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.VerifyImageRulesHandler(kyverno.NewPolicyStore()))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("No VerifyImage Rule Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/verify-image-rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		result := &kyverno.Rule{
			ValidateMessage: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
			Name:            "autogen-check-for-requests-and-limits",
		}

		policy := kyverno.Policy{
			Kind:              "Policy",
			Name:              "require-ressources",
			Namespace:         "test",
			Rules:             []*kyverno.Rule{result},
			CreationTimestamp: time.Now(),
		}

		store := kyverno.NewPolicyStore()
		store.Add(policy)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.VerifyImageRulesHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
	t.Run("VerifyImageRule Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/verify-image-rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		result := &kyverno.Rule{
			Name: "check-image",
			VerifyImages: []*kyverno.VerifyImage{
				{
					Image:      "ghcr.io/kyverno/test-verify-image:*",
					Repository: "registry.io/signatures",
					Key: `
					-----BEGIN PUBLIC KEY-----
					MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8nXRh950IZbRj8Ra/N9sbqOPZrfM
					5/KAQN0/KjHcorm/J5yctVd7iEcnessRQjU917hmKO6JWVGHpDguIyakZA==
					-----END PUBLIC KEY----- 
					`,
				},
			},
		}

		policy := kyverno.Policy{
			Kind:              "Policy",
			Name:              "check-image",
			Namespace:         "test",
			Rules:             []*kyverno.Rule{result},
			CreationTimestamp: time.Now(),
		}

		store := kyverno.NewPolicyStore()
		store.Add(policy)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.VerifyImageRulesHandler(store))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `{"policy":{"name":"check-image","namespace":"test"},"rule":"check-image","repository":"registry.io/signatures","image":"ghcr.io/kyverno/test-verify-image:*","key":"\n\t\t\t\t\t-----BEGIN PUBLIC KEY-----\n\t\t\t\t\tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8nXRh950IZbRj8Ra/N9sbqOPZrfM\n\t\t\t\t\t5/KAQN0/KjHcorm/J5yctVd7iEcnessRQjU917hmKO6JWVGHpDguIyakZA==\n\t\t\t\t\t-----END PUBLIC KEY----- \n\t\t\t\t\t"}`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}

func Test_PolicyReportingHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/policy-details-reporting", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.PolicyReportingHandler(&policyReportGeneratorStub{}, path.Join("..", "..", "templates", "reporting")))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("handler returned wrong status code: got %v want %v (%s)", status, http.StatusOK, string(b))
	}
}

func Test_PolicyReportingHandlerDataError(t *testing.T) {
	req, err := http.NewRequest("GET", "/policy-details-reporting", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.PolicyReportingHandler(&policyReportGeneratorStub{errors.New("error")}, path.Join("..", "..", "templates", "reporting")))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func Test_NamespaceReportingHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/namespace-details-reporting", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.NamespaceReportingHandler(&policyReportGeneratorStub{}, path.Join("..", "..", "templates", "reporting")))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("handler returned wrong status code: got %v want %v (%s)", status, http.StatusOK, string(b))
	}
}

func Test_NamespaceReportingHandlerDataError(t *testing.T) {
	req, err := http.NewRequest("GET", "/namespace-details-reporting", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.NamespaceReportingHandler(&policyReportGeneratorStub{errors.New("error")}, path.Join("..", "..", "templates", "reporting")))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
