package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
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

func (c *Client) FetchExternalContributions(ctx context.Context, username string) (domain.User, []domain.ContributionEvent, error) {
	user, err := c.fetchUser(ctx, username)
	if err != nil {
		return domain.User{}, nil, err
	}
	events, err := c.FetchExternalPRs(ctx, username)
	return user, events, err
}

func (c *Client) fetchUser(ctx context.Context, username string) (domain.User, error) {
	var q struct {
		User struct {
			Login     string
			AvatarURL githubv4.URI `graphql:"avatarUrl"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login": githubv4.String(username),
	}
	if err := c.gv4.Query(ctx, &q, variables); err != nil {
		return domain.User{}, fmt.Errorf("fetching user info: %w", err)
	}
	return domain.User{
		Username:  q.User.Login,
		AvatarURL: q.User.AvatarURL.String(),
	}, nil
}

func (c *Client) FetchExternalPRs(ctx context.Context, username string) ([]domain.ContributionEvent, error) {
	// Query: PRs authored by user, excluding their own repos
	query := fmt.Sprintf("author:%s type:pr -user:%s", username, username)
	return c.searchPRs(ctx, query)
}

func (c *Client) FetchOwnRepoPRs(ctx context.Context, username string, minStars int) ([]domain.ContributionEvent, error) {
	// Query: PRs authored by user, in their own repos with > minStars
	query := fmt.Sprintf("author:%s type:pr user:%s stars:>%d", username, username, minStars)
	return c.searchPRs(ctx, query)
}

func (c *Client) FetchOwnedProjects(ctx context.Context, username string, minStars int) ([]domain.OwnedProject, error) {
	var projects []domain.OwnedProject
	variables := map[string]interface{}{
		"login":  githubv4.String(username),
		"cursor": (*githubv4.String)(nil),
	}

	for {
		var q ownedRepoQuery
		if err := c.gv4.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("listing owned repos: %w", err)
		}

		for _, repo := range q.User.Repositories.Nodes {
			if repo.IsPrivate || repo.IsFork {
				continue
			}
			if repo.StargazerCount < minStars {
				continue
			}

			projects = append(projects, domain.OwnedProject{
				Repo:  repo.NameWithOwner,
				URL:   repo.URL.String(),
				Stars: repo.StargazerCount,
				Forks: repo.ForkCount,
			})
		}

		if !q.User.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.User.Repositories.PageInfo.EndCursor)
	}

	return projects, nil
}

type ownedRepoQuery struct {
	User struct {
		Repositories struct {
			Nodes []struct {
				NameWithOwner  string
				URL            githubv4.URI
				StargazerCount int
				ForkCount      int
				IsFork         bool
				IsPrivate      bool
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first: 100, ownerAffiliations: OWNER, after: $cursor)"`
		AvatarURL githubv4.URI `graphql:"avatarUrl"`
	} `graphql:"user(login: $login)"`
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
					Owner          struct {
						AvatarURL githubv4.URI `graphql:"avatarUrl"`
					}
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

func (c *Client) searchPRs(ctx context.Context, queryStr string) ([]domain.ContributionEvent, error) {
	var allEvents []domain.ContributionEvent
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

	return allEvents, nil
}
