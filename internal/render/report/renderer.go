package report

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

type Report struct {
	GeneratedAt   time.Time                  `json:"generatedAt"`
	Username      string                     `json:"username"`
	Stats         domain.UserStats           `json:"stats"`
	TotalEvents   int                        `json:"totalEvents"`
	EventsByType  map[string]int             `json:"eventsByType"`
	Events        []domain.ContributionEvent `json:"events"`
	OwnedProjects []domain.OwnedProject      `json:"ownedProjects"`
}

func (Renderer) RenderReport(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	eventsByType := make(map[string]int)
	for _, event := range events {
		eventsByType[string(event.Type)]++
	}

	report := Report{
		GeneratedAt:   generatedAt,
		Username:      user.Username,
		Stats:         stats,
		TotalEvents:   len(events),
		EventsByType:  eventsByType,
		Events:        events,
		OwnedProjects: projects,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling report: %w", err)
	}

	return data, nil
}
