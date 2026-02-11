package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
	"github.com/arayofcode/footprint/internal/github"
	"github.com/arayofcode/footprint/internal/logic"
	"github.com/arayofcode/footprint/internal/render/assets"
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

	events = g.Scorer.ScoreBatch(events)
	enrichedProjects := enrichOwnedProjects(g.Scorer, projects)

	// Semantic Pipeline
	semanticEvents := logic.MapClassify(events)
	statsView, repoContribs, projectImpacts := logic.Aggregate(semanticEvents, enrichedProjects)

	// Projection adapter: Attach finalized contributions to repo summaries
	finalizedEvents := domain.MapEventsToContributions(semanticEvents)
	repoEvents := make(map[string][]domain.Contribution)
	for _, fe := range finalizedEvents {
		repoEvents[fe.Repo] = append(repoEvents[fe.Repo], fe)
	}
	for i := range repoContribs {
		repoContribs[i].Events = repoEvents[repoContribs[i].Repo]
	}

	generatedAt := time.Now()

	reportJSON, err := g.ReportRenderer.RenderReport(ctx, user, statsView, generatedAt, repoContribs, projectImpacts)
	if err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}

	summaryMD, err := g.SummaryRenderer.RenderSummary(ctx, user, statsView, generatedAt, repoContribs, projectImpacts)
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
		for _, r := range repoContribs {
			totalScore += r.Score
		}
		for _, p := range projectImpacts {
			totalScore += p.Score
		}

		g.Actions.SetOutput("total_contributions", fmt.Sprintf("%d", len(events)))
		g.Actions.SetOutput("owned_projects_count", fmt.Sprintf("%d", len(projectImpacts)))
		g.Actions.SetOutput("total_score", fmt.Sprintf("%.2f", totalScore))
	}

	if g.CardRenderer != nil {
		// Fetch assets (avatars)
		assetMap := assets.FetchAssets(user, repoContribs, projectImpacts)

		// Render cards with finalized impact
		cardSVG, err := g.CardRenderer.RenderCard(ctx, user, statsView, generatedAt, repoContribs, projectImpacts, assetMap)
		if err != nil {
			return fmt.Errorf("rendering card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card.svg", cardSVG); err != nil {
			return fmt.Errorf("writing card.svg: %w", err)
		}

		// Minimal card
		minimalSVG, err := g.CardRenderer.RenderMinimalCard(ctx, user, statsView, generatedAt, repoContribs, projectImpacts, assetMap)
		if err != nil {
			return fmt.Errorf("rendering minimal card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-minimal.svg", minimalSVG); err != nil {
			return fmt.Errorf("writing card-minimal.svg: %w", err)
		}

		// Extended card
		extendedSVG, err := g.CardRenderer.RenderExtendedCard(ctx, user, statsView, generatedAt, repoContribs, projectImpacts, assetMap)
		if err != nil {
			return fmt.Errorf("rendering extended card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-extended.svg", extendedSVG); err != nil {
			return fmt.Errorf("writing card-extended.svg: %w", err)
		}

		// Extended-minimal card
		extMinimalSVG, err := g.CardRenderer.RenderExtendedMinimalCard(ctx, user, statsView, generatedAt, repoContribs, projectImpacts, assetMap)
		if err != nil {
			return fmt.Errorf("rendering extended-minimal card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-extended-minimal.svg", extMinimalSVG); err != nil {
			return fmt.Errorf("writing card-extended-minimal.svg: %w", err)
		}
	}

	return nil
}

func enrichOwnedProjects(calculator domain.ScoreCalculator, projects []domain.OwnedProject) []domain.EnrichedProject {
	enriched := make([]domain.EnrichedProject, 0, len(projects))
	for _, project := range projects {
		enriched = append(enriched, calculator.EnrichOwnedProject(project))
	}
	return enriched
}
