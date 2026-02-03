package card

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestRenderCard_GeneratesSVGWithStats(t *testing.T) {
	renderer := Renderer{}
	events := []domain.ContributionEvent{
		{Type: domain.ContributionTypePR, Repo: "a/b", Merged: true, Score: 5},
		{Type: domain.ContributionTypeIssue, Repo: "a/b", Score: 2},
	}
	projects := []domain.OwnedProject{
		{Repo: "me/owned", Stars: 10, Forks: 2},
	}
	user := domain.User{Username: "ray", AvatarURL: "https://example.com/avatar.png"}
	stats := domain.UserStats{
		TotalPRs:           5,
		TotalReviews:       3,
		TotalIssues:        2,
		TotalIssueComments: 10,
	}

	out, err := renderer.RenderCard(context.Background(), user, stats, time.Now(), events, projects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Check for SVG header
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG output")
	}

	// Check for new stats labels
	expectedLabels := []string{
		"PRS OPENED",
		"PR REVIEWS",
		"ISSUES OPENED",
		"ISSUE COMMENTS",
		"PROJECTS OWNED",
		"STARS EARNED",
	}
	for _, label := range expectedLabels {
		if !strings.Contains(svg, label) {
			t.Errorf("expected SVG to contain label %q", label)
		}
	}

	// Check for values
	if !strings.Contains(svg, ">5<") { // Total PRs
		t.Error("expected SVG to contain PR count 5")
	}
	if !strings.Contains(svg, ">10<") { // Stars (from projects)
		t.Error("expected SVG to contain Star count 10")
	}

	// Check for owned section
	if !strings.Contains(svg, "Owned Projects") {
		t.Error("expected SVG to contain 'Owned Projects' section")
	}

	// Check for external repos section (from events)
	if !strings.Contains(svg, "Top Repositories") {
		t.Error("expected SVG to contain 'Top Repositories' section")
	}
}
