package logic

import (
	"testing"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestAggregate(t *testing.T) {
	projects := []domain.OwnedProject{
		{Repo: "me/owned", Stars: 10},
	}

	events := []domain.SemanticEvent{
		{Type: domain.SemanticEventPrOpened, Repo: "ext/repo", Score: 5, AvatarURL: "avatar1"},
		{Type: domain.SemanticEventPrReview, Repo: "ext/repo", Score: 3},
		{Type: domain.SemanticEventIssueOpened, Repo: "other/repo", Score: 2},
		{Type: domain.SemanticEventIssueComment, Repo: "me/owned", Score: 2}, // Should be excluded from contributions
	}

	stats, contribs := Aggregate(events, projects)

	// Verify StatsView
	if stats.ProjectsOwned != 1 {
		t.Errorf("expected 1 project owned, got %d", stats.ProjectsOwned)
	}
	if stats.StarsEarned != 10 {
		t.Errorf("expected 10 stars earned, got %d", stats.StarsEarned)
	}
	if stats.PRsOpened != 1 {
		t.Errorf("expected 1 PR opened, got %d", stats.PRsOpened)
	}
	if stats.PRReviews != 1 {
		t.Errorf("expected 1 PR Review, got %d", stats.PRReviews)
	}
	if stats.IssuesOpened != 1 {
		t.Errorf("expected 1 issue opened, got %d", stats.IssuesOpened)
	}
	if stats.IssueComments != 1 {
		t.Errorf("expected 1 issue comment, got %d", stats.IssueComments)
	}
	if stats.TotalReposContributedTo != 3 {
		t.Errorf("expected 3 repos contributed to (including owned), got %d", stats.TotalReposContributedTo)
	}

	// Verify RepoContributions
	// Should contain "ext/repo" and "other/repo", but NOT "me/owned"
	if len(contribs) != 2 {
		t.Errorf("expected 2 repo contributions, got %d", len(contribs))
	}

	for _, c := range contribs {
		if c.Repo == "me/owned" {
			t.Error("expected owned repo to be excluded from contributions")
		}
		if c.Repo == "ext/repo" {
			if c.Score != 8 { // 5 + 3
				t.Errorf("expected ext/repo score 8, got %f", c.Score)
			}
			if c.AvatarURL != "avatar1" {
				t.Errorf("expected ext/repo avatar avatar1, got %s", c.AvatarURL)
			}
		}
	}
}

func TestAggregate_DoesNotMixContributionTypes(t *testing.T) {
	events := []domain.SemanticEvent{
		{Type: domain.SemanticEventIssueComment, Repo: "a", Score: 1},
		{Type: domain.SemanticEventPrReviewComment, Repo: "a", Score: 1},
		{Type: domain.SemanticEventPrReview, Repo: "a", Score: 1},
	}
	projects := []domain.OwnedProject{}

	stats, contribs := Aggregate(events, projects)

	// verify stats
	if stats.IssueComments != 1 {
		t.Errorf("Stats: Expected 1 IssueComment, got %d", stats.IssueComments)
	}
	if stats.PRReviewComments != 1 {
		t.Errorf("Stats: Expected 1 PRReviewComment, got %d", stats.PRReviewComments)
	}
	if stats.PRReviews != 1 {
		t.Errorf("Stats: Expected 1 PRReview, got %d", stats.PRReviews)
	}

	// verify contribs
	if len(contribs) != 1 {
		t.Fatalf("Expected 1 contribution, got %d", len(contribs))
	}
	c := contribs[0]
	if c.IssueComments != 1 {
		t.Errorf("Contrib: Expected 1 IssueComment, got %d", c.IssueComments)
	}
	if c.PRReviewComments != 1 {
		t.Errorf("Contrib: Expected 1 PRReviewComment, got %d", c.PRReviewComments)
	}
	if c.PRReviews != 1 {
		t.Errorf("Contrib: Expected 1 PRReview, got %d", c.PRReviews)
	}
}
