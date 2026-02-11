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

func (Renderer) RenderSummary(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, projects []domain.RepoContribution, ownedProjects []domain.OwnedProjectImpact) ([]byte, error) {
	_ = ctx

	var sb strings.Builder

	fmt.Fprintf(&sb, "# OSS Footprint: @%s\n\n", user.Username)
	fmt.Fprintf(&sb, "*Generated on %s*\n\n", generatedAt.Format("January 2, 2006"))

	sb.WriteString("## Impact Snapshot\n\n")
	fmt.Fprintf(&sb, "- 🔀 **%d** PRs Opened\n", stats.PRsOpened)
	fmt.Fprintf(&sb, "- 📋 **%d** PR Reviews\n", stats.PRReviews+stats.PRReviewComments)
	fmt.Fprintf(&sb, "- 🐛 **%d** Issues Opened\n", stats.IssuesOpened)
	fmt.Fprintf(&sb, "- 💬 **%d** Issue Comments\n", stats.IssueComments)
	fmt.Fprintf(&sb, "- 📦 **%d** Projects Owned\n", stats.ProjectsOwned)
	fmt.Fprintf(&sb, "- ⭐ **%s** Stars Earned\n\n", formatLargeNum(stats.StarsEarned))
	fmt.Fprintf(&sb, "[View all external PRs authored by @%s](https://github.com/pulls?q=is%%3Apr+author%%3A%s+-user%%3A%s)\n\n", user.Username, user.Username, user.Username)

	if len(ownedProjects) > 0 {
		sb.WriteString("## Owned Projects\n\n")
		sort.Slice(ownedProjects, func(i, j int) bool {
			if ownedProjects[i].Score != ownedProjects[j].Score {
				return ownedProjects[i].Score > ownedProjects[j].Score
			}
			return ownedProjects[i].Repo < ownedProjects[j].Repo
		})
		for _, project := range ownedProjects {
			fmt.Fprintf(&sb, "- [`%s`](%s) · ⭐ %s · 🍴 %s\n", project.Repo, project.URL, formatLargeNum(project.Stars), formatLargeNum(project.Forks))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Top Repositories\n\n")

	// Sort projects (external contributions) by Score desc
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Score != projects[j].Score {
			return projects[i].Score > projects[j].Score
		}
		return projects[i].Repo < projects[j].Repo
	})

	for _, p := range projects {
		repo := p.Repo
		fmt.Fprintf(&sb, "### [%s](https://github.com/%s/pulls?q=is%%3Apr+author%%3A%s)\n\n", repo, repo, user.Username)
		fmt.Fprintf(&sb, "*Total Impact: **%.1f** · %d PR(s)*\n\n", p.Score, p.PRsOpened)

		// Finalized events are already chronological or can be sorted here
		events := p.Events
		sort.Slice(events, func(i, j int) bool {
			return events[i].CreatedAt.After(events[j].CreatedAt)
		})

		for _, event := range events {
			sb.WriteString(formatOutputEvent(event))
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String()), nil
}

func formatOutputEvent(event domain.Contribution) string {
	date := event.CreatedAt.Format("Jan 2, 2006")

	var icon string
	switch event.Type {
	case domain.ContributionPR:
		icon = "🔀"
	case domain.ContributionPRReview:
		icon = "👀"
	case domain.ContributionPRReviewComment:
		icon = "💭"
	case domain.ContributionIssue:
		icon = "🐛"
	case domain.ContributionIssueComment:
		icon = "💬"
	case domain.ContributionDiscussion:
		icon = "💡"
	case domain.ContributionDiscussionComment:
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

func formatLargeNum(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}
