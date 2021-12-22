package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

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

		policy := &kyverno.Policy{
			Kind:              "Policy",
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

		expected := `[{"kind":"Policy","name":"require-ressources","namespace":"test","background":false,"rules":[{"message":"validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/","name":"autogen-check-for-requests-and-limits","type":""}]`
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
		handler := http.HandlerFunc(api.HealthzHandler(map[string]bool{"Policy": true}))

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
		handler := api.HealthzHandler(map[string]bool{})

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

		policy := &kyverno.Policy{
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

		policy := &kyverno.Policy{
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
