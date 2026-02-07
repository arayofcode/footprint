package logic

import (
	"github.com/arayofcode/footprint/internal/domain"
)

func Aggregate(events []domain.SemanticEvent, projects []domain.OwnedProject) (domain.StatsView, []domain.RepoContribution) {
	stats := domain.StatsView{
		ProjectsOwned: len(projects),
	}

	for _, p := range projects {
		stats.StarsEarned += p.Stars
	}

	repoMap := make(map[string]*domain.RepoContribution)
	uniqueRepos := make(map[string]bool)

	for _, e := range events {
		uniqueRepos[e.Repo] = true

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
		}

		// Skip own repos for contribution list
		isOwned := false
		for _, p := range projects {
			if e.Repo == p.Repo {
				isOwned = true
				break
			}
		}
		if isOwned {
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
		contrib.Score += e.Score

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
		}
	}

	stats.TotalReposContributedTo = len(uniqueRepos)

	var contributions []domain.RepoContribution
	for _, c := range repoMap {
		contributions = append(contributions, *c)
	}

	return stats, contributions
}
