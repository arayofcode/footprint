package report

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestRenderReport_BasicFields(t *testing.T) {
	renderer := Renderer{}
	projects := []domain.RepoContribution{
		{
			Repo:  "a/b",
			Score: 10.5,
			Events: []domain.Contribution{
				{Type: domain.ContributionPR, Repo: "a/b", CreatedAt: time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)},
				{Type: domain.ContributionIssue, Repo: "a/b", CreatedAt: time.Date(2025, 2, 1, 11, 0, 0, 0, time.UTC)},
			},
		},
	}
	ownedProjects := []domain.OwnedProjectImpact{
		{Repo: "me/owned", URL: "https://github.com/me/owned", Stars: 10, Forks: 2, Score: 3.5},
	}
	generatedAt := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	user := domain.User{Username: "ray", AvatarURL: "https://avatar.com/ray"}

	out, err := renderer.RenderReport(context.Background(), user, domain.StatsView{}, generatedAt, projects, ownedProjects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var report Report
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}

	if report.Username != "ray" {
		t.Fatalf("expected username ray, got %s", report.Username)
	}
	if !report.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("expected generatedAt %v, got %v", generatedAt, report.GeneratedAt)
	}
	if report.TotalEvents != 2 {
		t.Fatalf("expected totalEvents 2, got %d", report.TotalEvents)
	}
	if report.EventsByType["PR"] != 1 || report.EventsByType["ISSUE"] != 1 {
		t.Fatalf("unexpected eventsByType counts: %+v", report.EventsByType)
	}
	if len(report.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(report.Events))
	}
	if len(report.OwnedProjects) != 1 {
		t.Fatalf("expected 1 owned project, got %d", len(report.OwnedProjects))
	}
}

func TestRenderReport_EventsByTypeEmptyWhenNoEvents(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}

	out, err := renderer.RenderReport(context.Background(), user, domain.StatsView{}, time.Now(), nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var report Report
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}

	if report.TotalEvents != 0 {
		t.Fatalf("expected totalEvents 0, got %d", report.TotalEvents)
	}
	if len(report.EventsByType) != 0 {
		t.Fatalf("expected eventsByType empty, got %+v", report.EventsByType)
	}
}
