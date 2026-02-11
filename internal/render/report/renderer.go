package report

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

type Report struct {
	GeneratedAt    time.Time                   `json:"generatedAt"`
	Username       string                      `json:"username"`
	Stats          domain.StatsView            `json:"stats"`
	TotalEvents    int                         `json:"totalEvents"`
	EventsByType   map[string]int              `json:"eventsByType"`
	Events         []domain.Contribution       `json:"events"`
	OwnedProjects  []domain.OwnedProjectImpact `json:"ownedProjects"`
	TopRepos       []RepoImpact                `json:"topRepos"`
	ExternalPRsURL string                      `json:"externalPRsUrl"`
}

type RepoImpact struct {
	Repo        string  `json:"repo"`
	RepoURL     string  `json:"repoURL"`
	ImpactScore float64 `json:"impactScore"`
	PRCount     int     `json:"prCount"`
}

func (Renderer) RenderReport(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, projects []domain.RepoContribution, ownedProjects []domain.OwnedProjectImpact) ([]byte, error) {
	_ = ctx

	eventsByType := make(map[string]int)
	var allFinalEvents []domain.Contribution
	var topRepos []RepoImpact

	for _, p := range projects {
		repoURL := "https://github.com/" + p.Repo
		topRepos = append(topRepos, RepoImpact{
			Repo:        p.Repo,
			RepoURL:     repoURL,
			ImpactScore: p.Score,
			PRCount:     p.PRsOpened,
		})

		for _, e := range p.Events {
			allFinalEvents = append(allFinalEvents, e)
			eventsByType[string(e.Type)]++
		}
	}

	// Sort topRepos by ImpactScore
	sort.Slice(topRepos, func(i, j int) bool {
		return topRepos[i].ImpactScore > topRepos[j].ImpactScore
	})

	// Sort ownedProjects by final Score
	sort.Slice(ownedProjects, func(i, j int) bool {
		if ownedProjects[i].Score != ownedProjects[j].Score {
			return ownedProjects[i].Score > ownedProjects[j].Score
		}
		return ownedProjects[i].Repo < ownedProjects[j].Repo
	})

	// Sort events by date (desc)
	sort.Slice(allFinalEvents, func(i, j int) bool {
		return allFinalEvents[i].CreatedAt.After(allFinalEvents[j].CreatedAt)
	})

	report := Report{
		GeneratedAt:    generatedAt,
		Username:       user.Username,
		Stats:          stats,
		TotalEvents:    len(allFinalEvents),
		EventsByType:   eventsByType,
		Events:         allFinalEvents,
		OwnedProjects:  ownedProjects,
		TopRepos:       topRepos,
		ExternalPRsURL: fmt.Sprintf("https://github.com/pulls?q=is:pr+author:%s+-user:%s", user.Username, user.Username),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling report: %w", err)
	}

	return data, nil
}
