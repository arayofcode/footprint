package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type IssueAuthoredStrategy struct {
	client *githubv4.Client
}

func NewIssueAuthoredStrategy(client *githubv4.Client) *IssueAuthoredStrategy {
	return &IssueAuthoredStrategy{client: client}
}

func (s *IssueAuthoredStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	query := fmt.Sprintf("author:%s -user:%s type:issue", username, username)
	events, _, err := searchWithCount(ctx, s.client, query)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *IssueAuthoredStrategy) Name() domain.ContributionType {
	return domain.ContributionTypeIssue
}
