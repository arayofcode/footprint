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
	projects := []domain.RepoContribution{
		{
			Repo:      "a/b",
			Score:     10.0,
			PRsOpened: 1,
			Events: []domain.Contribution{
				{Type: domain.ContributionPR, Repo: "a/b", Merged: true, CreatedAt: time.Now()},
			},
		},
	}
	ownedProjects := []domain.OwnedProjectImpact{
		{Repo: "me/owned", URL: "https://github.com/me/owned", Stars: 10, Forks: 2, Score: 3.5},
	}
	generatedAt := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.StatsView{PRsOpened: 1, ProjectsOwned: 1, IssuesOpened: 1}, generatedAt, projects, ownedProjects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content := string(out)
	assertContains(t, content, "# OSS Footprint: @ray")
	assertContains(t, content, "Generated on February 1, 2025")
	assertContains(t, content, "## Impact Snapshot")
	assertContains(t, content, "🔀 **1** PRs Opened")
	assertContains(t, content, "🐛 **1** Issues Opened")
	assertContains(t, content, "## Owned Projects")
	assertContains(t, content, "`me/owned`")
}

func TestRenderSummary_FormatsRepositorySection(t *testing.T) {
	renderer := Renderer{}
	projects := []domain.RepoContribution{
		{
			Repo:      "a/b",
			Score:     16.7,
			PRsOpened: 1,
			Events: []domain.Contribution{
				{
					Type:      domain.ContributionPR,
					Repo:      "a/b",
					URL:       "https://github.com/a/b/pull/1",
					Title:     "Fix bug",
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					Type:      domain.ContributionIssue,
					Repo:      "a/b",
					URL:       "https://github.com/a/b/issues/2",
					Title:     "File issue",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.StatsView{}, time.Now(), projects, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content := string(out)
	assertContains(t, content, "## Top Repositories")
	assertContains(t, content, "### [a/b](https://github.com/a/b/pulls?q=is%3Apr+author%3Aray)")
	assertContains(t, content, "*Total Impact: **16.7** · 1 PR(s)*")
	assertContains(t, content, "**[Fix bug](https://github.com/a/b/pull/1)**")
	assertContains(t, content, "**[File issue](https://github.com/a/b/issues/2)**")
}

func TestRenderSummary_IncludesReactionsAndMergedFlags(t *testing.T) {
	renderer := Renderer{}
	projects := []domain.RepoContribution{
		{
			Repo:  "a/b",
			Score: 20.0,
			Events: []domain.Contribution{
				{
					Type:           domain.ContributionPR,
					Repo:           "a/b",
					URL:            "https://github.com/a/b/pull/1",
					Title:          "Add feature",
					CreatedAt:      time.Now(),
					ReactionsCount: 3,
					Merged:         true,
				},
			},
		},
	}
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderSummary(context.Background(), user, domain.StatsView{}, time.Now(), projects, nil)
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
