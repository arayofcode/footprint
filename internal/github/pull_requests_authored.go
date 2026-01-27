package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type PullRequestAuthoredStrategy struct {
	client *githubv4.Client
}

func NewPullRequestAuthoredStrategy(client *githubv4.Client) *PullRequestAuthoredStrategy {
	return &PullRequestAuthoredStrategy{client: client}
}

func (s *PullRequestAuthoredStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	query := fmt.Sprintf("author:%s type:pr -user:%s", username, username)
	events, _, err := searchPRsWithCount(ctx, s.client, query)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PullRequestAuthoredStrategy) Name() domain.ContributionType {
	return domain.ContributionTypePR
}
