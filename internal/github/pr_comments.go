package github

import (
	"context"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type PullRequestCommentStrategy struct {
	client *githubv4.Client
}

func NewPullRequestCommentStrategy(client *githubv4.Client) *PullRequestCommentStrategy {
	return &PullRequestCommentStrategy{client: client}
}

func (s *PullRequestCommentStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	return searchPullRequestComments(ctx, s.client, username)
}

func (s *PullRequestCommentStrategy) Name() domain.ContributionType {
	return domain.ContributionTypePRComment
}
