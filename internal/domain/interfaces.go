package domain

import (
	"context"
	"time"
)

type EventFetcher interface {
	FetchExternalContributions(ctx context.Context, username string) (User, []ContributionEvent, error)
}

type ContributionStrategy interface {
	Fetch(ctx context.Context, username string) ([]ContributionEvent, error)
	Name() ContributionType
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
	Bio       string `json:"bio,omitempty"`
	Company   string `json:"company,omitempty"`
	Location  string `json:"location,omitempty"`
	Website   string `json:"website,omitempty"`
	Followers int    `json:"followers_count,omitempty"`
}

type ReportRenderer interface {
	RenderReport(ctx context.Context, user User, stats UserStats, generatedAt time.Time, events []ContributionEvent, projects []OwnedProject) ([]byte, error)
}

type SummaryRenderer interface {
	RenderSummary(ctx context.Context, user User, stats UserStats, generatedAt time.Time, events []ContributionEvent, projects []OwnedProject) ([]byte, error)
}

type CardRenderer interface {
	RenderCard(ctx context.Context, user User, stats StatsView, generatedAt time.Time, contributions []RepoContribution, projects []OwnedProject) ([]byte, error)
	RenderMinimalCard(ctx context.Context, user User, stats StatsView, generatedAt time.Time, contributions []RepoContribution, projects []OwnedProject) ([]byte, error)
	RenderExtendedCard(ctx context.Context, user User, stats StatsView, generatedAt time.Time, contributions []RepoContribution, projects []OwnedProject) ([]byte, error)
	RenderExtendedMinimalCard(ctx context.Context, user User, stats StatsView, generatedAt time.Time, contributions []RepoContribution, projects []OwnedProject) ([]byte, error)
}

type OutputWriter interface {
	Write(ctx context.Context, filename string, data []byte) error
}
