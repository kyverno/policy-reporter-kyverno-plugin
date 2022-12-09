package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

const port int = 9999

func Test_NewServer(t *testing.T) {
	server := api.NewServer(kyverno.NewPolicyStore(), &policyReportGeneratorStub{}, port, func() bool { return true })

	server.RegisterMetrics()
	server.RegisterREST()

	serviceRunning := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		close(serviceRunning)
		err := server.Start()
		if err != nil {
			fmt.Println(err)
		}
		defer close(serviceDone)
	}()

	<-serviceRunning

	client := http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ready", port), nil)
	res, _ := client.Do(req)

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected Error Code: %d", res.StatusCode)
	}

	server.Shutdown(context.Background())

	<-serviceDone
}
