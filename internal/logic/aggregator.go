package logic

import (
	"github.com/arayofcode/footprint/internal/domain"
)

const RepoMultiplierCap = 4.0

// Aggregate reduces semantic events and enriched projects into finalized impact summaries.
// Logic:
// 1. External Impact: Sum BaseScore per repo, take Max(PopularityRaw), apply cap, multiply.
// 2. Owned Projects: Use BaseScore, apply cap to PopularityRaw, multiply.
// 3. Stats: Sum raw activity counts (unweighted).
func Aggregate(events []domain.SemanticEvent, projects []domain.EnrichedProject) (domain.StatsView, []domain.RepoContribution, []domain.OwnedProjectImpact) {
	var stats domain.StatsView
	repoMap := make(map[string]*domain.RepoContribution)

	// Track all unique repos for stats
	allRepos := make(map[string]bool)
	ownedRepos := make(map[string]bool)
	for _, p := range projects {
		allRepos[p.Repo] = true
		ownedRepos[p.Repo] = true
	}

	// First pass: Calculate overall stats and identify unique repos
	for _, e := range events {
		allRepos[e.Repo] = true

		// Update StatsView (Counts only)
		switch e.Type {
		case domain.SemanticEventPrOpened:
			stats.PRsOpened++
		case domain.SemanticEventPrReview:
			stats.PRReviews++
		case domain.SemanticEventPrReviewComment:
			stats.PRReviewComments++
		case domain.SemanticEventIssueOpened:
			stats.IssuesOpened++
		case domain.SemanticEventIssueComment:
			stats.IssueComments++
		case domain.SemanticEventDiscussionOpened:
			stats.DiscussionsOpened++
		case domain.SemanticEventDiscussionComment:
			stats.DiscussionComments++
		case domain.SemanticEventCommit:
			stats.TotalCommits++
		}

		// Skip owned projects for the external contributions breakdown
		if ownedRepos[e.Repo] {
			continue
		}

		if _, ok := repoMap[e.Repo]; !ok {
			repoMap[e.Repo] = &domain.RepoContribution{
				Repo:      e.Repo,
				RepoURL:   "https://github.com/" + e.Repo,
				AvatarURL: e.AvatarURL,
			}
		}

		contrib := repoMap[e.Repo]
		contrib.BaseScore += e.BaseScore
		if e.PopularityRaw > contrib.PopularityRaw {
			contrib.PopularityRaw = e.PopularityRaw
		}

		// Contribution projection moved to separate adapter

		switch e.Type {
		case domain.SemanticEventPrOpened:
			contrib.PRsOpened++
		case domain.SemanticEventPrReview:
			contrib.PRReviews++
		case domain.SemanticEventPrReviewComment:
			contrib.PRReviewComments++
		case domain.SemanticEventIssueOpened:
			contrib.IssuesOpened++
		case domain.SemanticEventIssueComment:
			contrib.IssueComments++
		case domain.SemanticEventDiscussionOpened:
			contrib.DiscussionsOpened++
		case domain.SemanticEventDiscussionComment:
			contrib.DiscussionComments++
		case domain.SemanticEventCommit:
			contrib.Commits++
		}
	}
	stats.TotalReposContributedTo = len(allRepos)

	var contributions []domain.RepoContribution
	for _, c := range repoMap {
		// Apply capped popularity multiplier at repo level
		multiplier := c.PopularityRaw
		if multiplier > RepoMultiplierCap {
			multiplier = RepoMultiplierCap
		}
		if multiplier < 1.0 {
			multiplier = 1.0
		}
		c.Score = c.BaseScore * multiplier // Final weighted score

		contributions = append(contributions, *c)
	}

	// Finalize Owned Projects
	var projectImpacts []domain.OwnedProjectImpact
	for _, p := range projects {
		stats.ProjectsOwned++
		stats.StarsEarned += p.Stars

		multiplier := p.PopularityRaw
		if multiplier > RepoMultiplierCap {
			multiplier = RepoMultiplierCap
		}
		if multiplier < 1.0 {
			multiplier = 1.0
		}

		projectImpacts = append(projectImpacts, domain.OwnedProjectImpact{
			Repo:          p.Repo,
			URL:           p.URL,
			AvatarURL:     p.AvatarURL,
			Stars:         p.Stars,
			Forks:         p.Forks,
			BaseScore:     p.BaseScore,
			PopularityRaw: p.PopularityRaw,
			Score:         p.BaseScore * multiplier,
		})
	}

	return stats, contributions, projectImpacts
}
