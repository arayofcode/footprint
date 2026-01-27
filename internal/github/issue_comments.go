package github

import (
	"context"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type IssueCommentsStrategy struct {
	client *githubv4.Client
}

func NewIssueCommentsStrategy(client *githubv4.Client) *IssueCommentsStrategy {
	return &IssueCommentsStrategy{client: client}
}

func (s *IssueCommentsStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	events, err := searchIssueComments(ctx, s.client, username)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *IssueCommentsStrategy) Name() domain.ContributionType {
	return domain.ContributionTypeIssueComment
}
