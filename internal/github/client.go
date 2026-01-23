package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v81/github"
)

type Client struct {
	gh *github.Client
}

func NewClient(ghClient *github.Client) *Client {
	if ghClient == nil {
		ghClient = github.NewClient(nil)
	}
	return &Client{gh: ghClient}
}

func (c *Client) FetchExternalPRs(ctx context.Context, username string) ([]*ContributionEvent, error) {
	// "author:username type:pr -user:username"
	query := fmt.Sprintf("author:%s type:pr -user:%s", username, username)

	var allEvents []*ContributionEvent
	opts := &github.SearchOptions{
		Sort:  "created",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		result, resp, err := c.gh.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("searching PRs: %w", err)
		}

		for _, issue := range result.Issues {
			event := issueToContributionEvent(issue, ContributionTypePR)
			allEvents = append(allEvents, event)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allEvents, nil
}

func (c *Client) FetchOwnRepoPRs(ctx context.Context, username string, minStars int) ([]*ContributionEvent, error) {
	popularRepos, err := c.getPopularRepos(ctx, username, minStars)
	if err != nil {
		return nil, err
	}

	if len(popularRepos) == 0 {
		return nil, nil
	}

	// "author:username type:pr repo:owner/repo1 repo:owner/repo2 ..."
	var repoFilters []string
	for _, repo := range popularRepos {
		repoFilters = append(repoFilters, fmt.Sprintf("repo:%s", repo))
	}

	query := fmt.Sprintf("author:%s type:pr %s", username, strings.Join(repoFilters, " "))

	var allEvents []*ContributionEvent
	opts := &github.SearchOptions{
		Sort:  "created",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		result, resp, err := c.gh.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("searching own repo PRs: %w", err)
		}

		for _, issue := range result.Issues {
			event := issueToContributionEvent(issue, ContributionTypePR)
			allEvents = append(allEvents, event)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allEvents, nil
}

// return repo full names (owner/repo) for user's repos with >= minStars.
func (c *Client) getPopularRepos(ctx context.Context, username string, minStars int) ([]string, error) {
	var popularRepos []string
	opts := &github.RepositoryListByUserOptions{
		Type:      "owner",
		Sort:      "pushed",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.gh.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("listing user repos: %w", err)
		}

		for _, repo := range repos {
			if repo.GetStargazersCount() > minStars {
				popularRepos = append(popularRepos, repo.GetFullName())
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return popularRepos, nil
}

func issueToContributionEvent(issue *github.Issue, eventType ContributionType) *ContributionEvent {
	event := &ContributionEvent{
		ID:        fmt.Sprintf("pr-%d", issue.GetID()),
		Type:      eventType,
		URL:       issue.GetHTMLURL(),
		Title:     issue.GetTitle(),
		CreatedAt: issue.GetCreatedAt().Time,
	}

	if url := issue.GetHTMLURL(); url != "" {
		event.Repo = extractRepoFromURL(url)
	}
	if reactions := issue.GetReactions(); reactions != nil {
		event.ReactionsCount = reactions.GetTotalCount()
	}

	// for detecting PR merging status
	if issue.GetState() == "closed" && issue.GetPullRequestLinks() != nil {
		// Note: To accurately detect merged status, we'd need to fetch the PR details
		// For now, we mark closed PRs - merged detection requires additional API call
		event.Merged = false
	}

	return event
}

func extractRepoFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 5 {
		return parts[3] + "/" + parts[4]
	}
	return ""
}
