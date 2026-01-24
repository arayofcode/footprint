package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

type Client struct {
	gv4 *githubv4.Client
}

func NewClient(gv4Client *githubv4.Client) *Client {
	return &Client{
		gv4: gv4Client,
	}
}

func (c *Client) FetchExternalPRs(ctx context.Context, username string) ([]*ContributionEvent, error) {
	// Query: PRs authored by user, excluding their own repos
	query := fmt.Sprintf("author:%s type:pr -user:%s", username, username)
	return c.searchPRs(ctx, query)
}

func (c *Client) FetchOwnRepoPRs(ctx context.Context, username string, minStars int) ([]*ContributionEvent, error) {
	// Query: PRs authored by user, in their own repos with > minStars
	query := fmt.Sprintf("author:%s type:pr user:%s stars:>%d", username, username, minStars)
	return c.searchPRs(ctx, query)
}

type prSearchQuery struct {
	Search struct {
		Nodes []struct {
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
				}
				Reactions struct {
					TotalCount int
				} `graphql:"reactions(content: THUMBS_UP)"` // Just counting thumbs up for now as a proxy, or use total if available
			} `graphql:"... on PullRequest"`
		}
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $cursor)"`
}

func (c *Client) searchPRs(ctx context.Context, queryStr string) ([]*ContributionEvent, error) {
	var allEvents []*ContributionEvent
	variables := map[string]interface{}{
		"query":  githubv4.String(queryStr),
		"cursor": (*githubv4.String)(nil),
	}

	for {
		var q prSearchQuery
		err := c.gv4.Query(ctx, &q, variables)
		if err != nil {
			return nil, fmt.Errorf("graphql search error: %w", err)
		}

		for _, node := range q.Search.Nodes {
			pr := node.PullRequest
			event := &ContributionEvent{
				ID:             pr.ID,
				Type:           ContributionTypePR,
				Repo:           pr.Repository.NameWithOwner,
				URL:            pr.URL,
				Title:          pr.Title,
				CreatedAt:      pr.CreatedAt.Time,
				Stars:          pr.Repository.StargazerCount,
				Forks:          pr.Repository.ForkCount,
				Merged:         pr.Merged,
				ReactionsCount: pr.Reactions.TotalCount,
			}
			allEvents = append(allEvents, event)
		}

		if !q.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Search.PageInfo.EndCursor)
	}

	return allEvents, nil
}
