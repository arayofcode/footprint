package github

import (
	"context"
	"fmt"
	"time"

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
		},
	}
}

func (c *Client) FetchExternalContributions(ctx context.Context, username string) (domain.User, domain.UserStats, []domain.ContributionEvent, error) {
	user, allTimePRs, allTimeRepos, err := c.fetchUser(ctx, username)
	if err != nil {
		return domain.User{}, domain.UserStats{}, nil, err
	}

	// Fetch global stats. We loop back 5 years to get a better 'footprint'
	var stats domain.UserStats
	now := time.Now()
	for i := range 5 {
		yearStats, err := c.fetchUserStatsForYear(ctx, username, now.AddDate(-i, 0, 0))
		if err == nil {
			stats.TotalCommits += yearStats.TotalCommits
			stats.TotalIssues += yearStats.TotalIssues
			stats.TotalPRs += yearStats.TotalPRs
			stats.TotalReviews += yearStats.TotalReviews
			if yearStats.TotalReposCount > stats.TotalReposCount {
				stats.TotalReposCount = yearStats.TotalReposCount
			}
		}
	}

	// Fetch contributions using strategies
	eventMap := make(map[string]domain.ContributionEvent)
	uniqueRepos := make(map[string]bool)

	for _, strategy := range c.strategies {
		events, err := strategy.Fetch(ctx, username)
		if err != nil {
			// Log error but continue with other strategies?
			// For now, let's just log or ignore, or maybe partial failure is okay.
			// The original code treated errors in search as "nil" result (sometimes).
			// But let's be robust.
			continue
		}

		// Update stats if this strategy found more items than global stats
		// (Legacy logic adaptation: originally we updated stats.TotalPRs based on search count)
		// Here we only know the count of events found.
		if strategy.Name() == domain.ContributionTypePR {
			if len(events) > stats.TotalPRs {
				stats.TotalPRs = len(events)
			}
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

	// Apply all-time totals if they are higher (most robust source)
	if allTimePRs > stats.TotalPRs {
		stats.TotalPRs = allTimePRs
	}
	if allTimeRepos > stats.TotalReposCount {
		stats.TotalReposCount = allTimeRepos
	}

	// Ensure the repo count covers the unique repos we found in search
	if len(uniqueRepos) > stats.TotalReposCount {
		stats.TotalReposCount = len(uniqueRepos)
	}

	allEvents := make([]domain.ContributionEvent, 0, len(eventMap))
	for _, e := range eventMap {
		allEvents = append(allEvents, e)
	}

	return user, stats, allEvents, nil
}

func (c *Client) fetchUserStatsForYear(ctx context.Context, username string, date time.Time) (domain.UserStats, error) {
	var q struct {
		User struct {
			ContributionsCollection struct {
				TotalCommitContributions            int
				TotalIssueContributions             int
				TotalPullRequestContributions       int
				TotalPullRequestReviewContributions int
				RepositoryContributions             struct {
					TotalCount int
				} `graphql:"repositoryContributions(first: 1)"`
			} `graphql:"contributionsCollection(from: $from, to: $to)"`
		} `graphql:"user(login: $login)"`
	}

	// Calculate year range
	year := date.Year()
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	variables := map[string]any{
		"login": githubv4.String(username),
		"from":  githubv4.DateTime{Time: from},
		"to":    githubv4.DateTime{Time: to},
	}
	if err := c.gv4.Query(ctx, &q, variables); err != nil {
		return domain.UserStats{}, err
	}

	coll := q.User.ContributionsCollection
	return domain.UserStats{
		TotalCommits:    coll.TotalCommitContributions,
		TotalPRs:        coll.TotalPullRequestContributions,
		TotalIssues:     coll.TotalIssueContributions,
		TotalReviews:    coll.TotalPullRequestReviewContributions,
		TotalReposCount: coll.RepositoryContributions.TotalCount,
	}, nil
}

func (c *Client) fetchUser(ctx context.Context, username string) (domain.User, int, int, error) {
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
			PullRequests struct {
				TotalCount int
			} `graphql:"pullRequests(first: 1)"`
			RepositoriesContributedTo struct {
				TotalCount int
			} `graphql:"repositoriesContributedTo(first: 1)"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]any{
		"login": githubv4.String(username),
	}
	if err := c.gv4.Query(ctx, &q, variables); err != nil {
		return domain.User{}, 0, 0, fmt.Errorf("fetching user info: %w", err)
	}
	return domain.User{
		Username:  q.User.Login,
		AvatarURL: q.User.AvatarURL.String(),
		Bio:       q.User.Bio,
		Company:   q.User.Company,
		Location:  q.User.Location,
		Followers: q.User.Followers.TotalCount,
	}, q.User.PullRequests.TotalCount, q.User.RepositoriesContributedTo.TotalCount, nil
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
