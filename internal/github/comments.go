package github

import (
	"context"
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/shurcooL/githubv4"
)

type issueCommentSearchQuery struct {
	User struct {
		IssueComments struct {
			Nodes []struct {
				Typename   githubv4.String `graphql:"__typename"`
				ID         string
				URL        string
				CreatedAt  githubv4.DateTime
				Repository struct {
					NameWithOwner  string
					StargazerCount int
					ForkCount      int
					Owner          struct {
						AvatarURL githubv4.URI `graphql:"avatarUrl"`
					}
				}
				Issue struct {
					Title string
				}
				Reactions struct {
					TotalCount int
				} `graphql:"reactions(content: THUMBS_UP)"`
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"issueComments(first: 100, after: $cursor)"`
	} `graphql:"user(login: $username)"`
}

func searchIssueComments(ctx context.Context, client *githubv4.Client, username string) ([]domain.ContributionEvent, error) {
	var allEvents []domain.ContributionEvent
	variables := map[string]any{
		"username": githubv4.String(username),
		"cursor":   (*githubv4.String)(nil),
	}

	for {
		var q issueCommentSearchQuery
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return nil, fmt.Errorf("graphql search error: %w", err)
		}

		for _, node := range q.User.IssueComments.Nodes {
			event := domain.ContributionEvent{
				ID:                 node.ID,
				Type:               domain.ContributionTypeIssueComment,
				Repo:               node.Repository.NameWithOwner,
				URL:                node.URL,
				Title:              node.Issue.Title,
				CreatedAt:          node.CreatedAt.Time,
				Stars:              node.Repository.StargazerCount,
				Forks:              node.Repository.ForkCount,
				ReactionsCount:     node.Reactions.TotalCount,
				RepoOwnerAvatarURL: node.Repository.Owner.AvatarURL.String(),
			}
			allEvents = append(allEvents, event)
		}

		if !q.User.IssueComments.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.User.IssueComments.PageInfo.EndCursor)
	}

	return allEvents, nil
}

type pullRequestCommentSearchQuery struct {
	User struct {
		PullRequestComments struct {
			Nodes []struct {
				ID         string
				URL        string
				CreatedAt  githubv4.DateTime
				Repository struct {
					NameWithOwner  string
					StargazerCount int
					ForkCount      int
					Owner          struct {
						AvatarURL githubv4.URI `graphql:"avatarUrl"`
					}
				}
				PullRequest struct {
					Title string
				}
				Reactions struct {
					TotalCount int
				} `graphql:"reactions(content: THUMBS_UP)"`
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"pullRequestComments(first: 100, after: $cursor)"`
	} `graphql:"user(login: $username)"`
}

func searchPullRequestComments(ctx context.Context, client *githubv4.Client, username string) ([]domain.ContributionEvent, error) {
	var allEvents []domain.ContributionEvent
	variables := map[string]any{
		"username": githubv4.String(username),
		"cursor":   (*githubv4.String)(nil),
	}

	for {
		var q pullRequestCommentSearchQuery
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return nil, fmt.Errorf("graphql search error: %w", err)
		}

		for _, node := range q.User.PullRequestComments.Nodes {
			event := domain.ContributionEvent{
				ID:                 node.ID,
				Type:               domain.ContributionTypePRComment,
				Repo:               node.Repository.NameWithOwner,
				URL:                node.URL,
				Title:              node.PullRequest.Title,
				CreatedAt:          node.CreatedAt.Time,
				Stars:              node.Repository.StargazerCount,
				Forks:              node.Repository.ForkCount,
				ReactionsCount:     node.Reactions.TotalCount,
				RepoOwnerAvatarURL: node.Repository.Owner.AvatarURL.String(),
			}
			allEvents = append(allEvents, event)
		}

		if !q.User.PullRequestComments.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.User.PullRequestComments.PageInfo.EndCursor)
	}

	return allEvents, nil
}
