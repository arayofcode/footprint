package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type Client struct {
	gv4        *githubv4.Client
	strategies []domain.ContributionStrategy
}

func NewClient(gv4Client *githubv4.Client) *Client {
	return &Client{
		gv4: gv4Client,
		strategies: []domain.ContributionStrategy{
			NewPullRequestAuthoredStrategy(gv4Client),
			NewPullRequestReviewedStrategy(gv4Client),
			NewPullRequestCommentStrategy(gv4Client),
			NewIssueAuthoredStrategy(gv4Client),
			NewIssueCommentsStrategy(gv4Client),
		},
	}
}

func (c *Client) FetchExternalContributions(ctx context.Context, username string) (domain.User, domain.UserStats, []domain.ContributionEvent, error) {
	user, err := c.fetchUser(ctx, username)

	if err != nil {
		return domain.User{}, domain.UserStats{}, nil, err
	}

	// Fetch contributions using strategies
	eventMap := make(map[string]domain.ContributionEvent)
	uniqueRepos := make(map[string]bool)

	for _, strategy := range c.strategies {
		events, err := strategy.Fetch(ctx, username)
		if err != nil {
			continue
		}

		for _, e := range events {
			uniqueRepos[e.Repo] = true
			if _, ok := eventMap[e.ID]; ok {
				// preserve existing logic: reviews might overwrite or be handled specifically
				// For now, last writer wins, which mimics previous behavior if Review strategy is last
				// But we need to ensure the Type is correct. The strategy sets the Type.
				// So if we have duplicates, the last strategy's version key wins.
				eventMap[e.ID] = e
			} else {
				eventMap[e.ID] = e
			}
		}
	}

	allEvents := make([]domain.ContributionEvent, 0, len(eventMap))
	var stats domain.UserStats

	for _, e := range eventMap {
		allEvents = append(allEvents, e)
		switch e.Type {
		case domain.ContributionTypePR:
			stats.TotalPRs++
		case domain.ContributionTypeIssue:
			stats.TotalIssues++
		case domain.ContributionTypeReview:
			stats.TotalReviews++
		}
	}
	stats.TotalReposCount = len(uniqueRepos)

	return user, stats, allEvents, nil
}

func (c *Client) fetchUser(ctx context.Context, username string) (domain.User, error) {
	var q struct {
		User struct {
			Login     string
			AvatarURL githubv4.URI `graphql:"avatarUrl"`
			Bio       string
			Company   string
			Location  string
			Followers struct {
				TotalCount int
			}
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]any{
		"login": githubv4.String(username),
	}
	if err := c.gv4.Query(ctx, &q, variables); err != nil {
		return domain.User{}, fmt.Errorf("fetching user info: %w", err)
	}
	return domain.User{
		Username:  q.User.Login,
		AvatarURL: q.User.AvatarURL.String(),
		Bio:       q.User.Bio,
		Company:   q.User.Company,
		Location:  q.User.Location,
		Followers: q.User.Followers.TotalCount,
	}, nil
}

func (c *Client) FetchOwnedProjects(ctx context.Context, username string, minStars int) ([]domain.OwnedProject, error) {
	var projects []domain.OwnedProject
	variables := map[string]any{
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
				Repo:      repo.NameWithOwner,
				URL:       repo.URL.String(),
				AvatarURL: repo.Owner.AvatarURL.String(),
				Stars:     repo.StargazerCount,
				Forks:     repo.ForkCount,
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
				Owner          struct {
					AvatarURL githubv4.URI `graphql:"avatarUrl"`
				}
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first: 100, ownerAffiliations: OWNER, after: $cursor)"`
		AvatarURL githubv4.URI `graphql:"avatarUrl"`
	} `graphql:"user(login: $login)"`
}
