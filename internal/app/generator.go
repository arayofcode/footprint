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
	projects = scoreOwnedProjects(g.Scorer, projects)

	// Semantic Pipeline
	semanticEvents := logic.MapClassify(events)
	statsView, repoContribs := logic.Aggregate(semanticEvents, projects)

	// Map to legacy UserStats for older renderers
	stats := domain.UserStats{
		TotalPRs:     statsView.PRsOpened,
		TotalReviews: statsView.PRReviews,
		TotalIssues:  statsView.IssuesOpened,
		// Group PR review comments and issue comments into legacy TotalIssueComments?
		// Or keep them separate? UserStats has TotalIssueComments.
		// Let's sum them for now to maintain roughly similar volume metrics if that's what TotalIssueComments implied.
		// Or maybe TotalReviews should include review comments? User requested "PRReviewed must count reviews only".
		// So TotalReviews = PRReviews.
		// TotalIssueComments = IssueComments (and ignore PRReviewComments in legacy stats? or add them?)
		// Let's add them to TotalIssueComments to capture all "commentary".
		TotalIssueComments: statsView.IssueComments + statsView.PRReviewComments,
		TotalReposCount:    statsView.TotalReposContributedTo,
		TotalStarsEarned:   statsView.StarsEarned,
	}

	generatedAt := time.Now()

	reportJSON, err := g.ReportRenderer.RenderReport(ctx, user, stats, generatedAt, events, projects)
	if err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}

	summaryMD, err := g.SummaryRenderer.RenderSummary(ctx, user, stats, generatedAt, events, projects)
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

		g.Actions.SetOutput("total_contributions", fmt.Sprintf("%d", len(events)))    //nolint:errcheck
		g.Actions.SetOutput("owned_projects_count", fmt.Sprintf("%d", len(projects))) //nolint:errcheck
		g.Actions.SetOutput("total_score", fmt.Sprintf("%.2f", totalScore))           //nolint:errcheck
	}

	if g.CardRenderer != nil {
		// Fetch assets (avatars)
		assetMap := assets.FetchAssets(user, repoContribs, projects)

		// Used pre-calculated views
		cardSVG, err := g.CardRenderer.RenderCard(ctx, user, statsView, generatedAt, repoContribs, projects, assetMap)
		if err != nil {
			return fmt.Errorf("rendering card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card.svg", cardSVG); err != nil {
			return fmt.Errorf("writing card.svg: %w", err)
		}

		// Minimal card
		minimalSVG, err := g.CardRenderer.RenderMinimalCard(ctx, user, statsView, generatedAt, repoContribs, projects, assetMap)
		if err != nil {
			return fmt.Errorf("rendering minimal card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-minimal.svg", minimalSVG); err != nil {
			return fmt.Errorf("writing card-minimal.svg: %w", err)
		}

		// Extended card
		extendedSVG, err := g.CardRenderer.RenderExtendedCard(ctx, user, statsView, generatedAt, repoContribs, projects, assetMap)
		if err != nil {
			return fmt.Errorf("rendering extended card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-extended.svg", extendedSVG); err != nil {
			return fmt.Errorf("writing card-extended.svg: %w", err)
		}

		// Extended-minimal card
		extMinimalSVG, err := g.CardRenderer.RenderExtendedMinimalCard(ctx, user, statsView, generatedAt, repoContribs, projects, assetMap)
		if err != nil {
			return fmt.Errorf("rendering extended-minimal card: %w", err)
		}
		if err := g.Writer.Write(ctx, "card-extended-minimal.svg", extMinimalSVG); err != nil {
			return fmt.Errorf("writing card-extended-minimal.svg: %w", err)
		}
	}

	return nil
}

func scoreOwnedProjects(calculator domain.ScoreCalculator, projects []domain.OwnedProject) []domain.OwnedProject {
	scored := make([]domain.OwnedProject, 0, len(projects))
	for _, project := range projects {
		scored = append(scored, calculator.ScoreOwnedProject(project))
	}
	return scored
}
