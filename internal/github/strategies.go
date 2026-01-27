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

type PullRequestReviewedStrategy struct {
	client *githubv4.Client
}

func NewPullRequestReviewedStrategy(client *githubv4.Client) *PullRequestReviewedStrategy {
	return &PullRequestReviewedStrategy{client: client}
}

func (s *PullRequestReviewedStrategy) Fetch(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	query := fmt.Sprintf("reviewer:%s type:pr -user:%s", username, username)
	events, _, err := searchPRsWithCount(ctx, s.client, query)
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

// --- Helper Types & Functions ---

type prSearchQuery struct {
	Search struct {
		IssueCount int
		Nodes      []struct {
			PullRequest struct {
				ID         string
				Title      string
				URL        string
				CreatedAt  githubv4.DateTime
				State      githubv4.PullRequestState
				Merged     bool
				Repository struct {
					NameWithOwner  string
					StargazerCount int
					ForkCount      int
					Owner          struct {
						AvatarURL githubv4.URI `graphql:"avatarUrl"`
					}
				}
				Reactions struct {
					TotalCount int
				} `graphql:"reactions(content: THUMBS_UP)"`
			} `graphql:"... on PullRequest"`
		}
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $cursor)"`
}

func searchPRsWithCount(ctx context.Context, client *githubv4.Client, queryStr string) ([]domain.ContributionEvent, int, error) {
	var allEvents []domain.ContributionEvent
	totalCount := 0
	variables := map[string]any{
		"query":  githubv4.String(queryStr),
		"cursor": (*githubv4.String)(nil),
	}

	for {
		var q prSearchQuery
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return nil, 0, fmt.Errorf("graphql search error: %w", err)
		}

		totalCount = q.Search.IssueCount
		for _, node := range q.Search.Nodes {
			pr := node.PullRequest
			event := domain.ContributionEvent{
				ID:                 pr.ID,
				Type:               domain.ContributionTypePR,
				Repo:               pr.Repository.NameWithOwner,
				URL:                pr.URL,
				Title:              pr.Title,
				CreatedAt:          pr.CreatedAt.Time,
				Stars:              pr.Repository.StargazerCount,
				Forks:              pr.Repository.ForkCount,
				Merged:             pr.Merged,
				ReactionsCount:     pr.Reactions.TotalCount,
				RepoOwnerAvatarURL: pr.Repository.Owner.AvatarURL.String(),
			}
			allEvents = append(allEvents, event)
		}

		if !q.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Search.PageInfo.EndCursor)
	}

	return allEvents, totalCount, nil
}
