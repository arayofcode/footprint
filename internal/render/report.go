package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arayofcode/footprint/internal/github"
)

type Report struct {
	GeneratedAt  time.Time                   `json:"generatedAt"`
	Username     string                      `json:"username"`
	TotalEvents  int                         `json:"totalEvents"`
	EventsByType map[string]int              `json:"eventsByType"`
	Events       []*github.ContributionEvent `json:"events"`
}

func GenerateReport(username string, events []*github.ContributionEvent) *Report {
	eventsByType := make(map[string]int)
	for _, event := range events {
		eventsByType[string(event.Type)]++
	}

	return &Report{
		GeneratedAt:  time.Now(),
		Username:     username,
		TotalEvents:  len(events),
		EventsByType: eventsByType,
		Events:       events,
	}
}

func WriteReportJSON(report *Report, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling report: %w", err)
	}

	outputPath := filepath.Join(outputDir, "report.json")
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing report.json: %w", err)
	}

	return nil
}
