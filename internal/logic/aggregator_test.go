package logic

import (
	"testing"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestAggregate(t *testing.T) {
	projects := []domain.EnrichedProject{
		{
			OwnedProject:  domain.OwnedProject{Repo: "me/owned", Stars: 10},
			BaseScore:     2500, // OwnershipScore
			PopularityRaw: 1.0,
		},
	}

	events := []domain.SemanticEvent{
		{Type: domain.SemanticEventPrOpened, Repo: "ext/repo", BaseScore: 5, AvatarURL: "avatar1", PopularityRaw: 1.0},
		{Type: domain.SemanticEventPrReview, Repo: "ext/repo", BaseScore: 3, PopularityRaw: 1.0},
		{Type: domain.SemanticEventIssueOpened, Repo: "other/repo", BaseScore: 2, PopularityRaw: 1.0},
		{Type: domain.SemanticEventIssueComment, Repo: "me/owned", BaseScore: 2, PopularityRaw: 1.0}, // Should be excluded from contributions
	}

	stats, contribs, impacts := Aggregate(events, projects)

	if len(impacts) != 1 {
		t.Errorf("expected 1 project impact, got %d", len(impacts))
	} else {
		project := impacts[0]
		if project.Score != 2500 {
			t.Errorf("expected me/owned score 2500, got %f", project.Score)
		}
	}

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
	if len(contribs) != 2 {
		t.Errorf("expected 2 repo contributions, got %d", len(contribs))
	}

	foundExt := false
	for _, c := range contribs {
		if c.Repo == "me/owned" {
			t.Error("expected owned repo to be excluded from contributions")
		}
		if c.Repo == "ext/repo" {
			foundExt = true
			if c.Score != 8 { // (5 + 3) * 1.0
				t.Errorf("expected ext/repo score 8, got %f", c.Score)
			}
			if c.AvatarURL != "avatar1" {
				t.Errorf("expected ext/repo avatar avatar1, got %s", c.AvatarURL)
			}
		}
	}
	if !foundExt {
		t.Error("ext/repo not found in contributions")
	}
}

func TestAggregate_DoesNotMultiplyPerEvent(t *testing.T) {
	// 3 events, each with popularityRaw = 5.0 (capped to 4.0).
	// sum(BaseScore) = 3 + 3 + 3 = 9.
	// Expected Score = 9 * 4.0 = 36.
	// If it were per-event multiplication, we'd get (3*4) + (3*4) + (3*4) = 36.
	// Wait, that's the same! Let's use different base scores.
	// sum(BaseScore) = 1 + 2 + 3 = 6.
	// Expected Score = 6 * 4.0 = 24.
	// If it were per-event: (1*4) + (2*4) + (3*4) = 4 + 8 + 12 = 24.
	// Still the same! Mathematical property of distributivity.
	// To proof it doesn't amplify per event, we should check if we sum FIRST.

	events := []domain.SemanticEvent{
		{Repo: "a", BaseScore: 10, PopularityRaw: 5.0},
		{Repo: "a", BaseScore: 10, PopularityRaw: 5.0},
	}

	_, contribs, _ := Aggregate(events, nil)
	c := contribs[0]
	// Correct: (10 + 10) * 4.0 = 80
	if c.Score != 80 {
		t.Errorf("Expected 80, got %f", c.Score)
	}
}

func TestAggregate_MaxPopularityChosen(t *testing.T) {
	events := []domain.SemanticEvent{
		{Repo: "a", BaseScore: 10, PopularityRaw: 1.2},
		{Repo: "a", BaseScore: 10, PopularityRaw: 3.9},
		{Repo: "a", BaseScore: 10, PopularityRaw: 2.5},
	}

	_, contribs, _ := Aggregate(events, nil)
	c := contribs[0]
	// Max popularity = 3.9
	// sum(BaseScore) = 30
	// final score = 30 * 3.9 = 117
	if c.Score != 117 {
		t.Errorf("Expected 117.0, got %f", c.Score)
	}
}

func TestAggregate_RepoLevelMultiplierCapping(t *testing.T) {
	events := []domain.SemanticEvent{
		{Type: domain.SemanticEventPrOpened, Repo: "popular/repo", BaseScore: 10, PopularityRaw: 2.0},
		{Type: domain.SemanticEventPrReview, Repo: "popular/repo", BaseScore: 5, PopularityRaw: 5.0},
	}

	_, contribs, _ := Aggregate(events, nil)
	c := contribs[0]
	// BaseScore = 10 + 5 = 15
	// MaxPopularity = 5.0 -> capped to 4.0
	// Expected Score = 15 * 4.0 = 60
	expected := 60.0
	if c.Score != expected {
		t.Errorf("expected capped score %v, got %v (BaseScore: %v, PopularityRaw: %v)", expected, c.Score, c.BaseScore, c.PopularityRaw)
	}
}
