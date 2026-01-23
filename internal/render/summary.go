package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arayofcode/footprint/internal/github"
)

func WriteSummaryMarkdown(report *Report, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# OSS Footprint: @%s\n\n", report.Username))
	sb.WriteString(fmt.Sprintf("*Generated on %s*\n\n", report.GeneratedAt.Format("January 2, 2006")))
	sb.WriteString(fmt.Sprintf("**Total Contributions:** %d\n\n", report.TotalEvents))

	if len(report.EventsByType) > 0 {
		sb.WriteString("### By Type\n\n")
		for eventType, count := range report.EventsByType {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", eventType, count))
		}
		sb.WriteString("\n")
	}

	repoGroups := groupByRepo(report.Events)

	sb.WriteString("## Contributions by Repository\n\n")
	type repoCount struct {
		repo  string
		count int
	}
	var repos []repoCount
	for repo, events := range repoGroups {
		repos = append(repos, repoCount{repo, len(events)})
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].count > repos[j].count
	})

	for _, rc := range repos {
		repo := rc.repo
		events := repoGroups[repo]

		sb.WriteString(fmt.Sprintf("### [`%s`](https://github.com/%s)\n\n", repo, repo))
		sb.WriteString(fmt.Sprintf("*%d contribution(s)*\n\n", len(events)))

		sort.Slice(events, func(i, j int) bool {
			return events[i].CreatedAt.After(events[j].CreatedAt)
		})

		for _, event := range events {
			sb.WriteString(formatEvent(event))
		}
		sb.WriteString("\n")
	}

	outputPath := filepath.Join(outputDir, "summary.md")
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing summary.md: %w", err)
	}

	return nil
}

func groupByRepo(events []*github.ContributionEvent) map[string][]*github.ContributionEvent {
	groups := make(map[string][]*github.ContributionEvent)
	for _, event := range events {
		groups[event.Repo] = append(groups[event.Repo], event)
	}
	return groups
}

func formatEvent(event *github.ContributionEvent) string {
	date := event.CreatedAt.Format("Jan 2, 2006")

	var icon string
	switch event.Type {
	case github.ContributionTypePR:
		icon = "🔀"
	case github.ContributionTypeIssue:
		icon = "🐛"
	case github.ContributionTypeIssueComment:
		icon = "💬"
	case github.ContributionTypeReview:
		icon = "👀"
	case github.ContributionTypeReviewComment:
		icon = "💭"
	case github.ContributionTypeDiscussion:
		icon = "💡"
	case github.ContributionTypeDiscussionComment:
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
