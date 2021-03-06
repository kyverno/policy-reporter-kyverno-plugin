package api_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_GzipCompression(t *testing.T) {
	t.Run("GzipRespose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		handler := api.Gzip(api.PolicyHandler(kyverno.NewPolicyStore()))

		handler.ServeHTTP(rr, req)

		reader, err := gzip.NewReader(rr.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if buf.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", buf.String(), expected)
		}
	})
	t.Run("Uncompressed Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.Gzip(api.PolicyHandler(kyverno.NewPolicyStore()))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := `[]`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}
