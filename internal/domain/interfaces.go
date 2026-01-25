package domain

import (
	"context"
	"time"
)

type EventFetcher interface {
	FetchExternalContributions(ctx context.Context, username string) (User, []ContributionEvent, error)
}

type ProjectCatalog interface {
	FetchOwnedProjects(ctx context.Context, username string, minStars int) ([]OwnedProject, error)
}

type ScoreCalculator interface {
	ScoreContribution(event ContributionEvent) ContributionEvent
	ScoreOwnedProject(project OwnedProject) OwnedProject
}

type User struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type ReportRenderer interface {
	RenderReport(ctx context.Context, user User, generatedAt time.Time, events []ContributionEvent, projects []OwnedProject) ([]byte, error)
}

type SummaryRenderer interface {
	RenderSummary(ctx context.Context, user User, generatedAt time.Time, events []ContributionEvent, projects []OwnedProject) ([]byte, error)
}

type CardRenderer interface {
	RenderCard(ctx context.Context, user User, generatedAt time.Time, events []ContributionEvent, projects []OwnedProject) ([]byte, error)
}

type OutputWriter interface {
	Write(ctx context.Context, filename string, data []byte) error
}
