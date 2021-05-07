package api_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_NewServer(t *testing.T) {
	server := api.NewServer(
		kyverno.NewPolicyStore(),
		8080,
	)

	go server.Start()
}
