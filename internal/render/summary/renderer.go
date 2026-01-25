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

func (Renderer) RenderSummary(ctx context.Context, user domain.User, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	var sb strings.Builder

	// Calculate high-level stats
	mergedPRs := 0
	uniqueRepos := make(map[string]bool)
	for _, e := range events {
		if e.Type == domain.ContributionTypePR && e.Merged {
			mergedPRs++
		}
		uniqueRepos[e.Repo] = true
	}

	fmt.Fprintf(&sb, "# OSS Footprint: @%s\n\n", user.Username)
	fmt.Fprintf(&sb, "*Generated on %s*\n\n", generatedAt.Format("January 2, 2006"))

	sb.WriteString("## Impact Snapshot\n\n")
	fmt.Fprintf(&sb, "- **%d** Merged Pull Requests\n", mergedPRs)
	fmt.Fprintf(&sb, "- **%d** Popular Projects Owned\n", len(projects))
	fmt.Fprintf(&sb, "- **%d** Unique Repositories Contributed To\n\n", len(uniqueRepos))

	if len(projects) > 0 {
		sb.WriteString("## Owned Projects\n\n")
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].Stars > projects[j].Stars
		})
		for _, project := range projects {
			fmt.Fprintf(&sb, "- [`%s`](%s) · ⭐ %s · 🍴 %s\n", project.Repo, project.URL, formatLargeNum(project.Stars), formatLargeNum(project.Forks))
		}
		sb.WriteString("\n")
	}

	repoGroups := groupByRepo(events)

	sb.WriteString("## Contributions by Repository\n\n")
	type repoSummary struct {
		repo   string
		merged int
		total  int
	}
	var repos []repoSummary
	for repo, repoEvents := range repoGroups {
		merged := 0
		for _, e := range repoEvents {
			if e.Merged {
				merged++
			}
		}
		repos = append(repos, repoSummary{repo: repo, merged: merged, total: len(repoEvents)})
	}
	// Sort by total contributions
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].total > repos[j].total
	})

	for _, rs := range repos {
		repo := rs.repo
		repoEvents := repoGroups[repo]

		fmt.Fprintf(&sb, "### [`%s`](https://github.com/%s)\n\n", repo, repo)
		fmt.Fprintf(&sb, "*%d contribution(s) (%d merged)*\n\n", rs.total, rs.merged)

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
