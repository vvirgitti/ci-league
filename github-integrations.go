package ci_league

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"regexp"
	"time"
)

type GithubIntegrationsService struct {
	idMappings map[string]string
	client     *github.Client
}

func NewGithubIntegrationsService(client *github.Client, idMappings map[string]string) *GithubIntegrationsService {
	return &GithubIntegrationsService{client: client, idMappings: idMappings}
}

func (g *GithubIntegrationsService) GetIntegrations(ctx context.Context, owner string, repos []string) (TeamIntegrations, error) {
	frequency, err := g.getCommitFrequency(ctx, owner, repos, g.idMappings)

	if err != nil {
		return nil, err
	}

	integrations := NewTeamIntegrations(frequency)
	return integrations, nil
}

var (
	coAuthorRegex = regexp.MustCompile(`Co-authored-by:.*<(.*)>`)
)

func ExtractCoAuthor(message string) string {
	matches := coAuthorRegex.FindStringSubmatch(message)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (g *GithubIntegrationsService) getCommitFrequency(ctx context.Context, owner string, repos []string, idMappings map[string]string) (map[Dev]int, error) {

	allCommits, err := g.getCommits(ctx, owner, repos...)
	if err != nil {
		return nil, err
	}

	commitFrequency := make(map[string]int)
	avatars := make(map[string]string)
	for _, commit := range allCommits {

		name := commit.GetCommit().GetAuthor().GetEmail()

		if alias, found := idMappings[name]; found {
			name = alias
		}

		if name != "" {
			commitFrequency[name]++
			avatars[name] = commit.GetAuthor().GetAvatarURL()
		}

		if coAuthor := ExtractCoAuthor(commit.GetCommit().GetMessage()); coAuthor != "" {
			if alias, found := idMappings[coAuthor]; found {
				coAuthor = alias
			}
			commitFrequency[coAuthor]++
		}
	}

	devs := make(map[Dev]int)

	for name, score := range commitFrequency {
		devs[Dev{
			Name:   name,
			Avatar: avatars[name],
		}] = score
	}

	return devs, nil
}

func (g *GithubIntegrationsService) getCommits(ctx context.Context, owner string, repos ...string) ([]*github.RepositoryCommit, error) {
	var allCommits []*github.RepositoryCommit

	for _, repo := range repos {
		options := github.CommitsListOptions{
			Since:       monday(),
			ListOptions: github.ListOptions{},
		}
		for {
			commits, response, err := g.client.Repositories.ListCommits(ctx, owner, repo, &options)

			if err != nil {
				return nil, fmt.Errorf("couldn't get commits, %s", err)
			}

			allCommits = append(allCommits, commits...)

			if response.NextPage == 0 {
				break
			}

			options.Page = response.NextPage
		}
	}
	return allCommits, nil
}

func monday() time.Time {
	date := time.Now()

	for date.Weekday() != time.Monday {
		date = date.Add(-1 * (time.Hour * 24))
	}

	year, month, day := date.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}
