package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/arayofcode/footprint/internal/github"
)

type Generator struct {
	Fetcher         domain.EventFetcher
	Projects        domain.ProjectCatalog
	Scorer          domain.ScoreCalculator
	ReportRenderer  domain.ReportRenderer
	SummaryRenderer domain.SummaryRenderer
	CardRenderer    domain.CardRenderer
	Writer          domain.OutputWriter
	Actions         *github.Actions
	MinStars        int
}

func (g *Generator) Run(ctx context.Context, username string) error {
	if g.Fetcher == nil || g.Projects == nil || g.Scorer == nil || g.ReportRenderer == nil || g.SummaryRenderer == nil || g.Writer == nil {
		return fmt.Errorf("generator dependencies are not fully configured")
	}

	user, events, err := g.Fetcher.FetchExternalContributions(ctx, username)
	if err != nil {
		return fmt.Errorf("fetching external contributions: %w", err)
	}

	projects, err := g.Projects.FetchOwnedProjects(ctx, username, g.MinStars)
	if err != nil {
		return fmt.Errorf("fetching owned projects: %w", err)
	}

	events = scoreContributions(g.Scorer, events)
	projects = scoreOwnedProjects(g.Scorer, projects)

	generatedAt := time.Now()

	reportJSON, err := g.ReportRenderer.RenderReport(ctx, user, generatedAt, events, projects)
	if err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}

	summaryMD, err := g.SummaryRenderer.RenderSummary(ctx, user, generatedAt, events, projects)
	if err != nil {
		return fmt.Errorf("rendering summary: %w", err)
	}

	if err := g.Writer.Write(ctx, "report.json", reportJSON); err != nil {
		return fmt.Errorf("writing report.json: %w", err)
	}

	if err := g.Writer.Write(ctx, "summary.md", summaryMD); err != nil {
		return fmt.Errorf("writing summary.md: %w", err)
	}

	if g.Actions != nil {
		if err := g.Actions.WriteSummary(summaryMD); err != nil {
			fmt.Printf("Warning: failed to write job summary: %v\n", err)
		}

		totalScore := 0.0
		for _, e := range events {
			totalScore += e.Score
		}
		for _, p := range projects {
			totalScore += p.Score
		}

		g.Actions.SetOutput("total_contributions", fmt.Sprintf("%d", len(events)))
		g.Actions.SetOutput("owned_projects_count", fmt.Sprintf("%d", len(projects)))
		g.Actions.SetOutput("total_score", fmt.Sprintf("%.2f", totalScore))
	}

	if g.CardRenderer != nil {
		cardSVG, err := g.CardRenderer.RenderCard(ctx, user, generatedAt, events, projects)
		if err != nil {
			return fmt.Errorf("rendering card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card.svg", cardSVG); err != nil {
			return fmt.Errorf("writing card.svg: %w", err)
		}
	}

	return nil
}

func scoreContributions(calculator domain.ScoreCalculator, events []domain.ContributionEvent) []domain.ContributionEvent {
	scored := make([]domain.ContributionEvent, 0, len(events))
	for _, event := range events {
		scored = append(scored, calculator.ScoreContribution(event))
	}
	return scored
}

func scoreOwnedProjects(calculator domain.ScoreCalculator, projects []domain.OwnedProject) []domain.OwnedProject {
	scored := make([]domain.OwnedProject, 0, len(projects))
	for _, project := range projects {
		scored = append(scored, calculator.ScoreOwnedProject(project))
	}
	return scored
}
