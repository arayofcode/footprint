package summary

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestRenderSummary_IncludesHeaderAndTotals(t *testing.T) {
	renderer := Renderer{}
	events := []domain.ContributionEvent{
		{Type: domain.ContributionTypePR, Repo: "a/b", Merged: true},
		{Type: domain.ContributionTypeIssue, Repo: "a/b"},
	}
	projects := []domain.OwnedProject{
		{Repo: "me/owned", URL: "https://github.com/me/owned", Stars: 10, Forks: 2, Score: 3.5},
	}
	generatedAt := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.UserStats{TotalPRs: 1, TotalReposCount: 1}, generatedAt, events, projects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content := string(out)
	assertContains(t, content, "# OSS Footprint: @ray")
	assertContains(t, content, "Generated on February 1, 2025")
	assertContains(t, content, "## Impact Snapshot")
	assertContains(t, content, "**1** Merged Pull Requests")
	assertContains(t, content, "## Owned Projects")
	assertContains(t, content, "`me/owned`")
}

func TestRenderSummary_FormatsRepositorySection(t *testing.T) {
	renderer := Renderer{}
	events := []domain.ContributionEvent{
		{
			Type:      domain.ContributionTypePR,
			Repo:      "a/b",
			URL:       "https://github.com/a/b/pull/1",
			Title:     "Fix bug",
			CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Score:     12.5,
		},
		{
			Type:      domain.ContributionTypeIssue,
			Repo:      "a/b",
			URL:       "https://github.com/a/b/issues/2",
			Title:     "File issue",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Score:     4.2,
		},
	}
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.UserStats{}, time.Now(), events, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content := string(out)
	assertContains(t, content, "## Contributions by Repository")
	assertContains(t, content, "### [`a/b`](https://github.com/a/b)")
	assertContains(t, content, "*2 contribution(s) (0 merged)*")
	assertContains(t, content, "**[Fix bug](https://github.com/a/b/pull/1)**")
	assertContains(t, content, "**[File issue](https://github.com/a/b/issues/2)**")
}

func TestRenderSummary_IncludesReactionsAndMergedFlags(t *testing.T) {
	renderer := Renderer{}
	events := []domain.ContributionEvent{
		{
			Type:           domain.ContributionTypePR,
			Repo:           "a/b",
			URL:            "https://github.com/a/b/pull/1",
			Title:          "Add feature",
			CreatedAt:      time.Now(),
			Score:          20,
			ReactionsCount: 3,
			Merged:         true,
		},
	}
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.UserStats{}, time.Now(), events, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content := string(out)
	assertContains(t, content, "❤️ 3")
	assertContains(t, content, "✅ Merged")
}

func assertContains(t *testing.T, content, expected string) {
	t.Helper()
	if !strings.Contains(content, expected) {
		t.Fatalf("expected content to contain %q", expected)
	}
}
