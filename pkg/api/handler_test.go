package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_PolicyReportAPI(t *testing.T) {
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

		result := kyverno.Rule{
			ValidateMessage: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
			Name:            "autogen-check-for-requests-and-limits",
		}

		policy := kyverno.Policy{
			Kind:              "Policy",
			Name:              "require-ressources",
			Namespace:         "test",
			Rules:             []kyverno.Rule{result},
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
