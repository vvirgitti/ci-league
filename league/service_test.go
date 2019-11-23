package league_test

import (
	"context"
	"github.com/quii/ci-league/github"
	"github.com/quii/ci-league/league"
	"io/ioutil"
	"testing"
)

func TestGithubIntegrationsService_GetIntegrations(t *testing.T) {
	commitService := github.NewService(github.NewClient("", ioutil.Discard))
	service := league.NewService(commitService, nil)
	_, err := service.GetStats(context.Background(), "quii", []string{"ci-league"})

	if err != nil {
		t.Fatalf("Failed to get integrations %s", err)
	}
}
