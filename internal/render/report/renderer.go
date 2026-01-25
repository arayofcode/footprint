package report

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

type Report struct {
	GeneratedAt    time.Time                  `json:"generatedAt"`
	Username       string                     `json:"username"`
	Stats          domain.UserStats           `json:"stats"`
	TotalEvents    int                        `json:"totalEvents"`
	EventsByType   map[string]int             `json:"eventsByType"`
	Events         []domain.ContributionEvent `json:"events"`
	OwnedProjects  []domain.OwnedProject      `json:"ownedProjects"`
	TopRepos       []RepoImpact               `json:"topRepos"`
	ExternalPRsURL string                     `json:"externalPRsUrl"`
}

type RepoImpact struct {
	Repo        string  `json:"repo"`
	ImpactScore float64 `json:"impactScore"`
	PRCount     int     `json:"prCount"`
}

func (Renderer) RenderReport(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	eventsByType := make(map[string]int)
	repoImpactMap := make(map[string]*RepoImpact)
	var externalEvents []domain.ContributionEvent

	for _, event := range events {
		// Only truly external
		parts := strings.Split(event.Repo, "/")
		if len(parts) < 2 || parts[0] == user.Username {
			continue
		}

		externalEvents = append(externalEvents, event)
		eventsByType[string(event.Type)]++

		if _, ok := repoImpactMap[event.Repo]; !ok {
			repoImpactMap[event.Repo] = &RepoImpact{Repo: event.Repo}
		}
		repoImpactMap[event.Repo].ImpactScore += event.Score
		if event.Type == domain.ContributionTypePR {
			repoImpactMap[event.Repo].PRCount++
		}
	}

	var topRepos []RepoImpact
	for _, ri := range repoImpactMap {
		topRepos = append(topRepos, *ri)
	}
	sort.Slice(topRepos, func(i, j int) bool {
		return topRepos[i].ImpactScore > topRepos[j].ImpactScore
	})

	// Sort projects
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Stars != projects[j].Stars {
			return projects[i].Stars > projects[j].Stars
		}
		return projects[i].Forks > projects[j].Forks
	})

	// Sort events by impact score
	sort.Slice(externalEvents, func(i, j int) bool {
		if externalEvents[i].Score != externalEvents[j].Score {
			return externalEvents[i].Score > externalEvents[j].Score
		}
		return externalEvents[i].CreatedAt.After(externalEvents[j].CreatedAt)
	})

	report := Report{
		GeneratedAt:    generatedAt,
		Username:       user.Username,
		Stats:          stats,
		TotalEvents:    len(externalEvents),
		EventsByType:   eventsByType,
		Events:         externalEvents,
		OwnedProjects:  projects,
		TopRepos:       topRepos,
		ExternalPRsURL: fmt.Sprintf("https://github.com/pulls?q=is:pr+author:%s+-user:%s", user.Username, user.Username),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling report: %w", err)
	}

	return data, nil
}
