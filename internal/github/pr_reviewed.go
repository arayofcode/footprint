package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type PullRequestReviewedStrategy struct {
	client *githubv4.Client
}

func NewPullRequestReviewedStrategy(client *githubv4.Client) *PullRequestReviewedStrategy {
	return &PullRequestReviewedStrategy{client: client}
}

func (s *PullRequestReviewedStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	query := fmt.Sprintf("reviewer:%s -user:%s type:pr", username, username)
	events, _, err := searchWithCount(ctx, s.client, query)
	if err != nil {
		return nil, err
	}
	// Mark events as reviews
	for i := range events {
		events[i].Type = domain.ContributionTypeReview
	}
	return events, nil
}

func (s *PullRequestReviewedStrategy) Name() domain.ContributionType {
	return domain.ContributionTypeReview
}
