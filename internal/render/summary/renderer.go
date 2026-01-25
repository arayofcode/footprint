package summary

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

func (Renderer) RenderSummary(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	// 1. Filter out self-owned repo events and calculate accurate external stats
	var externalEvents []domain.ContributionEvent
	externalRepos := make(map[string]bool)
	mergedPRs := 0
	for _, e := range events {
		parts := strings.Split(e.Repo, "/")
		if len(parts) < 2 || parts[0] == user.Username {
			continue
		}
		externalEvents = append(externalEvents, e)
		externalRepos[e.Repo] = true
		if e.Type == domain.ContributionTypePR && e.Merged {
			mergedPRs++
		}
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "# OSS Footprint: @%s\n\n", user.Username)
	fmt.Fprintf(&sb, "*Generated on %s*\n\n", generatedAt.Format("January 2, 2006"))

	sb.WriteString("## Impact Snapshot\n\n")
	fmt.Fprintf(&sb, "- **%d** Popular Projects Owned\n", len(projects))
	fmt.Fprintf(&sb, "- **%d** Unique Repositories Contributed To\n", len(externalRepos))
	fmt.Fprintf(&sb, "- **%d** PRs Merged\n\n", mergedPRs)
	fmt.Fprintf(&sb, "[View all external PRs authored by @%s](https://github.com/pulls?q=is%%3Apr+author%%3A%s+-user%%3A%s)\n\n", user.Username, user.Username, user.Username)

	if len(projects) > 0 {
		sb.WriteString("## Owned Projects\n\n")
		sort.Slice(projects, func(i, j int) bool {
			if projects[i].Stars != projects[j].Stars {
				return projects[i].Stars > projects[j].Stars
			}
			return projects[i].Forks > projects[j].Forks
		})
		for _, project := range projects {
			fmt.Fprintf(&sb, "- [`%s`](%s) · ⭐ %s · 🍴 %s\n", project.Repo, project.URL, formatLargeNum(project.Stars), formatLargeNum(project.Forks))
		}
		sb.WriteString("\n")
	}

	repoGroups := groupByRepo(externalEvents)

	sb.WriteString("## Top Repositories\n\n")
	type repoSummary struct {
		repo        string
		impactScore float64
		prCount     int
		total       int
	}
	var repos []repoSummary
	for repo, repoEvents := range repoGroups {
		score := 0.0
		prs := 0
		for _, e := range repoEvents {
			score += e.Score
			if e.Type == domain.ContributionTypePR {
				prs++
			}
		}
		repos = append(repos, repoSummary{
			repo:        repo,
			impactScore: score,
			prCount:     prs,
			total:       len(repoEvents),
		})
	}

	// Sort by Impact Score
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].impactScore != repos[j].impactScore {
			return repos[i].impactScore > repos[j].impactScore
		}
		return repos[i].repo < repos[j].repo
	})

	for _, rs := range repos {
		repo := rs.repo
		repoEvents := repoGroups[repo]

		fmt.Fprintf(&sb, "### [%s](https://github.com/%s/pulls?q=is%%3Apr+author%%3A%s)\n\n", repo, repo, user.Username)
		fmt.Fprintf(&sb, "*Total Impact: **%.1f** · %d PR(s)*\n\n", rs.impactScore, rs.prCount)

		sort.Slice(repoEvents, func(i, j int) bool {
			return repoEvents[i].CreatedAt.After(repoEvents[j].CreatedAt)
		})

		for _, event := range repoEvents {
			sb.WriteString(formatEvent(event))
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String()), nil
}

func formatLargeNum(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

func groupByRepo(events []domain.ContributionEvent) map[string][]domain.ContributionEvent {
	groups := make(map[string][]domain.ContributionEvent)
	for _, event := range events {
		groups[event.Repo] = append(groups[event.Repo], event)
	}
	return groups
}

func formatEvent(event domain.ContributionEvent) string {
	date := event.CreatedAt.Format("Jan 2, 2006")

	var icon string
	switch event.Type {
	case domain.ContributionTypePR:
		icon = "🔀"
	case domain.ContributionTypeIssue:
		icon = "🐛"
	case domain.ContributionTypeIssueComment:
		icon = "💬"
	case domain.ContributionTypeReview:
		icon = "👀"
	case domain.ContributionTypeReviewComment:
		icon = "💭"
	case domain.ContributionTypeDiscussion:
		icon = "💡"
	case domain.ContributionTypeDiscussionComment:
		icon = "🗨️"
	default:
		icon = "📝"
	}

	line := fmt.Sprintf("- %s **[%s](%s)** (%s)", icon, event.Title, event.URL, date)

	if event.ReactionsCount > 0 {
		line += fmt.Sprintf(" · ❤️ %d", event.ReactionsCount)
	}

	if event.Merged {
		line += " · ✅ Merged"
	}

	line += "\n"
	return line
}
