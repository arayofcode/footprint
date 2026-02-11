package card

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

// RenderCard: All stats, no sections
func (r Renderer) RenderCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProjectImpact, assets map[domain.AssetKey]string) ([]byte, error) {
	vm := buildViewModel(user, stats, generatedAt, contributions, projects, true, false, false)
	return renderSVG(vm, assets), nil
}

// RenderMinimalCard: Non-zero stats only, no sections
func (r Renderer) RenderMinimalCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProjectImpact, assets map[domain.AssetKey]string) ([]byte, error) {
	vm := buildViewModel(user, stats, generatedAt, contributions, projects, false, false, false)
	return renderSVG(vm, assets), nil
}

// RenderExtendedCard: All stats + both sections
func (r Renderer) RenderExtendedCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProjectImpact, assets map[domain.AssetKey]string) ([]byte, error) {
	vm := buildViewModel(user, stats, generatedAt, contributions, projects, true, true, false)
	return renderSVG(vm, assets), nil
}

// RenderExtendedMinimalCard: Non-zero stats + sections only if content exists
func (r Renderer) RenderExtendedMinimalCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProjectImpact, assets map[domain.AssetKey]string) ([]byte, error) {
	vm := buildViewModel(user, stats, generatedAt, contributions, projects, false, true, true)
	return renderSVG(vm, assets), nil
}

func buildViewModel(user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProjectImpact, showAllStats bool, showSections bool, minimalSections bool) CardViewModel {
	codeReview := stats.PRReviewComments + stats.PRReviews
	// 1. Build Stats
	potentialStats := []StatVM{
		{Label: "PRs Opened", Value: formatCount(stats.PRsOpened), Icon: iconPR, Raw: stats.PRsOpened},
		{Label: "Code Reviews", Value: formatCount(codeReview), Icon: iconReview, Raw: codeReview},
		{Label: "Issues Opened", Value: formatCount(stats.IssuesOpened), Icon: iconIssue, Raw: stats.IssuesOpened},
		{Label: "Issue Comments", Value: formatCount(stats.IssueComments), Icon: iconComment, Raw: stats.IssueComments},
		{Label: "Projects Owned", Value: formatCount(len(projects)), Icon: iconProject, Raw: len(projects)},
		{Label: "Stars Earned", Value: formatLargeNum(stats.StarsEarned), Icon: iconStar, Raw: stats.StarsEarned},
	}

	var activeStats []StatVM
	for _, s := range potentialStats {
		if showAllStats || s.Raw > 0 {
			s.Color = "#22c55e"
			activeStats = append(activeStats, s)
		}
	}

	// 2. Prepare Sections Data
	topOwned := projects
	if showSections {
		// Sort by Score desc (finalized weighted impact)
		sort.Slice(projects, func(i, j int) bool {
			if projects[i].Score != projects[j].Score {
				return projects[i].Score > projects[j].Score
			}
			return projects[i].Repo < projects[j].Repo
		})
		if len(topOwned) > 3 {
			topOwned = topOwned[:3]
		}
	}

	topExternal := contributions
	// Sort by Score desc, then Repo asc
	sort.Slice(topExternal, func(i, j int) bool {
		if topExternal[i].Score != topExternal[j].Score {
			return topExternal[i].Score > topExternal[j].Score
		}
		return topExternal[i].Repo < topExternal[j].Repo
	})
	if len(topExternal) > 3 {
		topExternal = topExternal[:3]
	}

	hasOwned := len(topOwned) > 0
	hasExternal := len(topExternal) > 0

	statCount := len(activeStats)

	// Determine Placement (Unified logic)
	// View model just sets the default intent (Horizontal) and Auto mode
	mode := LayoutAuto

	// 3. Decide Layout
	var layoutSections []SectionLayoutInput
	colIdx := 0
	if showSections {
		if !minimalSections || hasOwned {
			layoutSections = append(layoutSections, SectionLayoutInput{
				Rows:      len(topOwned),
				IsEmpty:   !hasOwned,
				Placement: StackHorizontal,
				Column:    colIdx,
			})
			colIdx++
		}
		if !minimalSections || hasExternal {
			layoutSections = append(layoutSections, SectionLayoutInput{
				Rows:      len(topExternal),
				IsEmpty:   !hasExternal,
				Placement: StackHorizontal,
				Column:    colIdx,
			})
			colIdx++
		}
	}

	layoutInput := LayoutInput{
		StatCount:    statCount,
		ShowAllStats: showAllStats,
		Mode:         mode,
		Sections:     layoutSections,
	}
	layout := DecideLayout(layoutInput)

	// 4. Position Stats
	for i := range activeStats {
		var x, y int
		if layout.IsVertical {
			x = 0
			y = i * layout.StatSpacing
		} else if !showAllStats && statCount == 4 {
			col := i % 2
			row := i / 2
			x = col * 350
			y = row * 90
		} else {
			col := i % 3
			row := i / 3
			x = col * 250
			y = row * 90
		}
		activeStats[i].X = x
		activeStats[i].Y = y
	}

	// 5. Build Sections
	var sections []SectionVM
	if showSections {
		// Owned Section
		if !minimalSections || hasOwned {
			var rows []SectionRowVM
			for _, p := range topOwned {
				rows = append(rows, SectionRowVM{
					Kind:      RowOwnedProject,
					Title:     truncate(p.Repo, 25),
					Link:      p.URL,
					AvatarKey: domain.RepoAvatarKey(p.Repo),
					Badges: []BadgeVM{
						{Icon: iconStar, Count: formatCount(p.Stars)},
					},
				})
			}
			sections = append(sections, SectionVM{
				Title:        "TOP PROJECTS CREATED",
				EmptyMessage: "No projects created yet",
				Rows:         rows,
			})
		}

		// External Section
		if !minimalSections || hasExternal {
			var rows []SectionRowVM
			for _, r := range topExternal {
				parts := strings.Split(r.Repo, "/")
				repoName := r.Repo
				ownerName := ""
				if len(parts) == 2 {
					ownerName = parts[0]
					repoName = parts[1]
				}

				badges := []BadgeVM{}
				if r.PRsOpened > 0 {
					link := fmt.Sprintf("https://github.com/%s/pulls?q=is%%3Apr+author%%3A%s", r.Repo, user.Username)
					badges = append(badges, BadgeVM{Count: fmt.Sprintf("%d", r.PRsOpened), Icon: iconPR, Link: link})
				}
				if r.PRReviews > 0 {
					link := fmt.Sprintf("https://github.com/%s/pulls?q=is%%3Apr+reviewed-by%%3A%s", r.Repo, user.Username)
					badges = append(badges, BadgeVM{Count: fmt.Sprintf("%d", r.PRReviews), Icon: iconReview, Link: link})
				}
				if r.IssuesOpened > 0 || r.IssueComments > 0 || r.PRReviewComments > 0 {
					link := fmt.Sprintf("https://github.com/%s/issues?q=commenter%%3A%s", r.Repo, user.Username)
					count := r.IssuesOpened + r.IssueComments + r.PRReviewComments
					badges = append(badges, BadgeVM{Count: fmt.Sprintf("%d", count), Icon: iconIssue, Link: link})
				}

				rows = append(rows, SectionRowVM{
					Kind:      RowExternalContribution,
					Title:     truncate(repoName, 15),
					Subtitle:  truncate(ownerName, 20),
					Link:      fmt.Sprintf("https://github.com/%s", r.Repo),
					AvatarKey: domain.RepoAvatarKey(r.Repo),
					Badges:    badges,
				})
			}
			sections = append(sections, SectionVM{
				Title:        "KEY CONTRIBUTIONS",
				EmptyMessage: "No contributions yet",
				Rows:         rows,
			})
		}
	}

	footer := FooterVM{
		Y:           layout.Height - 25,
		GeneratedAt: generatedAt.Format("02 Jan 2006"),
	}
	if !layout.IsVertical {
		footer.Attribution = `<a xlink:href="https://github.com/arayofcode/footprint" target="_blank"><text x="400" y="0" text-anchor="middle" font-family="system-ui, -apple-system, sans-serif" font-size="11" font-weight="600" fill="#22c55e" style="cursor: pointer;">Generated by Footprint</text></a>`
	}

	return CardViewModel{
		Width:      layout.Width,
		Height:     layout.Height,
		IsVertical: layout.IsVertical,
		Layout:     layout,
		User: UserVM{
			Username:  user.Username,
			AvatarKey: domain.UserAvatarKey(user.Username),
		},
		Stats:    activeStats,
		Sections: sections,
		Footer:   footer,
	}
}

// renderSVG composes the final SVG string from ViewModel
func renderSVG(vm CardViewModel, assetsMap map[domain.AssetKey]string) []byte {
	resolveAsset := func(key domain.AssetKey) string {
		if val, ok := assetsMap[key]; ok {
			return val
		}
		return ""
	}

	userAvatar := resolveAsset(vm.User.AvatarKey)
	defs := renderDefs()
	bg := fmt.Sprintf(`<rect width="%d" height="%d" rx="16" fill="#1a1a1a" />`, vm.Width, vm.Height)
	header := renderHeader(vm.User, userAvatar)

	statsContent := ""
	var statBoxes []string
	for _, s := range vm.Stats {
		statBoxes = append(statBoxes, renderStatBox(s.X, s.Y, s.Label, s.Value, s.Icon, s.Color))
	}
	if len(statBoxes) > 0 {
		statsContent = fmt.Sprintf(`
  <g transform="translate(40, %d)">
    %s
  </g>`, HeaderMargin, strings.Join(statBoxes, "\n    "))
	}

	sectionsContent := ""
	for i, sec := range vm.Sections {
		// ZIP: Semantic Section with Layout Geometry
		loc := vm.Layout.Sections[i]
		body := renderSection(sec, vm.Layout, resolveAsset)
		sectionsContent += fmt.Sprintf(`
  <g transform="translate(%d, %d)">
    %s
    %s
  </g>`,
			loc.X,
			loc.Y,
			renderSectionHeader(sec.Title),
			body,
		)
	}

	footer := renderFooter(vm.Footer, vm.Width)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  %s
  
  <!-- Background -->
  %s
  
  <!-- Header -->
  %s
  
%s
  %s
  %s
</svg>
`,
		vm.Width, vm.Height, vm.Width, vm.Height,
		defs,
		bg,
		header,
		statsContent,
		sectionsContent,
		footer,
	)

	return []byte(svg)
}

func renderSection(sec SectionVM, layout LayoutVM, assetResolver func(domain.AssetKey) string) string {
	if len(sec.Rows) == 0 {
		return renderEmptyState(sec.EmptyMessage)
	}

	var sb strings.Builder
	rowHeight := layout.RowHeight
	cardWidth := SectionWidth
	if layout.IsVertical {
		cardWidth = VerticalSectionWidth
	}

	for i, row := range sec.Rows {
		y := 35 + (i * rowHeight)
		avatar := assetResolver(row.AvatarKey)

		subtitleSVG := ""
		titleY := 28
		if row.Subtitle != "" {
			titleY = 18
			subtitleSVG = fmt.Sprintf(`<text x="42" y="32" font-family="system-ui, -apple-system, sans-serif" font-size="10" fill="#9ca3af">%s</text>`, html.EscapeString(row.Subtitle))
		}

		badgesSVG := ""
		if len(row.Badges) > 0 {
			switch row.Kind {
			case RowOwnedProject:
				badgesSVG = renderOwnedBadges(row.Badges[0], cardWidth)
			case RowExternalContribution:
				badgesSVG = renderExternalBadges(row.Badges, cardWidth)
			}
		}

		sb.WriteString(fmt.Sprintf(`
    <a xlink:href="%s" target="_blank">
      <g transform="translate(0, %d)">
        <rect width="%d" height="45" rx="10" fill="#1f2937" opacity="0.3" stroke="#374151" stroke-width="1"/>
        <g transform="translate(10, 10.5)">
          <image href="%s" width="24" height="24" clip-path="url(#repo-clip)" x="0" y="0"/>
        </g>
        <text x="42" y="%d" font-family="system-ui, -apple-system, sans-serif" font-size="13" font-weight="600" fill="white">%s</text>
        %s
        %s
      </g>
    </a>`,
			html.EscapeString(row.Link),
			y,
			cardWidth,
			avatar,
			titleY,
			html.EscapeString(row.Title),
			subtitleSVG,
			badgesSVG,
		))
	}
	return sb.String()
}

func renderDefs() string {
	return `<defs>
    <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="8" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    <clipPath id="avatar-clip">
      <circle cx="60" cy="45" r="20" />
    </clipPath>
    <clipPath id="repo-clip">
       <circle cx="12" cy="12" r="12" />
    </clipPath>
  </defs>`
}

func renderHeader(user UserVM, avatar string) string {
	return fmt.Sprintf(`
  <a xlink:href="https://github.com/%s" target="_blank">
    <g>
      <image href="%s" x="40" y="25" width="40" height="40" clip-path="url(#avatar-clip)" />
      <circle cx="60" cy="45" r="20" fill="none" stroke="#22c55e" stroke-width="2"/>
      <text x="95" y="52" font-family="system-ui, -apple-system, sans-serif" font-size="24" font-weight="600" fill="white">%s</text>
    </g>
  </a>`, user.Username, avatar, user.Username)
}

func renderFooter(footer FooterVM, width int) string {
	return fmt.Sprintf(`
  <g transform="translate(40, %d)">
    <text x="0" y="0" font-family="system-ui, -apple-system, sans-serif" font-size="11" fill="#6b7280">Range: All-time</text>
    %s
    <text x="%d" y="0" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="11" fill="#6b7280">Last Updated %s</text>
  </g>`, footer.Y, footer.Attribution, width-80, footer.GeneratedAt)
}

func formatCount(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000.0)
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}

func renderSmallIconBox(icon string) string {
	return fmt.Sprintf(`
      <rect width="24" height="24" rx="6" fill="#1f2937" opacity="0.6" stroke="#22c55e" stroke-width="1"/>
      <g transform="translate(4.5, 4.5) scale(0.6)" stroke="#22c55e" fill="none">
        %s
      </g>`, icon)
}

func formatLargeNum(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

func renderStatBox(x, y int, label, value, iconXML, color string) string {
	return fmt.Sprintf(`
    <g transform="translate(%d, %d)">
      <rect x="-5" y="-5" width="70" height="70" rx="14" fill="%s" opacity="0.15" filter="url(#glow)"/>
      <rect width="60" height="60" rx="12" fill="#1f2937" stroke="%s" stroke-width="1.5" opacity="0.9"/>
      <g transform="translate(18, 18)" fill="%s">
        %s
      </g>
      <text x="75" y="35" font-family="system-ui, -apple-system, sans-serif" font-size="32" font-weight="700" fill="white">%s</text>
      <text x="75" y="52" font-family="system-ui, -apple-system, sans-serif" font-size="10" font-weight="600" fill="#9ca3af" letter-spacing="0.5">%s</text>
    </g>`, x, y, color, color, color, iconXML, value, strings.ToUpper(label))
}

const (
	iconPR      = `<path d="M16 19.25a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0Zm-14.5 0a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0Zm0-14.5a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0ZM4.75 3a1.75 1.75 0 1 0 .001 3.501A1.75 1.75 0 0 0 4.75 3Zm0 14.5a1.75 1.75 0 1 0 .001 3.501A1.75 1.75 0 0 0 4.75 17.5Zm14.5 0a1.75 1.75 0 1 0 .001 3.501 1.75 1.75 0 0 0-.001-3.501Z"></path><path d="M13.405 1.72a.75.75 0 0 1 0 1.06L12.185 4h4.065A3.75 3.75 0 0 1 20 7.75v8.75a.75.75 0 0 1-1.5 0V7.75a2.25 2.25 0 0 0-2.25-2.25h-4.064l1.22 1.22a.75.75 0 0 1-1.061 1.06l-2.5-2.5a.75.75 0 0 1 0-1.06l2.5-2.5a.75.75 0 0 1 1.06 0ZM4.75 7.25A.75.75 0 0 1 5.5 8v8A.75.75 0 0 1 4 16V8a.75.75 0 0 1 .75-.75Z"></path>`
	iconReview  = `<path d="M10.3 6.74a.75.75 0 0 1-.04 1.06l-2.908 2.7 2.908 2.7a.75.75 0 1 1-1.02 1.1l-3.5-3.25a.75.75 0 0 1 0-1.1l3.5-3.25a.75.75 0 0 1 1.06.04Zm3.44 1.06a.75.75 0 1 1 1.02-1.1l3.5 3.25a.75.75 0 0 1 0 1.1l-3.5 3.25a.75.75 0 1 1-1.02-1.1l2.908-2.7-2.908-2.7Z"></path><path d="M1.5 4.25c0-.966.784-1.75 1.75-1.75h17.5c.966 0 1.75.784 1.75 1.75v12.5a1.75 1.75 0 0 1-1.75 1.75h-9.69l-3.573 3.573A1.458 1.458 0 0 1 5 21.043V18.5H3.25a1.75 1.75 0 0 1-1.75-1.75ZM3.25 4a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h2.5a.75.75 0 0 1 .75.75v3.19l3.72-3.72a.749.749 0 0 1 .53-.22h10a.25.25 0 0 0 .25-.25V4.25a.25.25 0 0 0-.25-.25Z"></path>`
	iconIssue   = `<path d="M12 1c6.075 0 11 4.925 11 11s-4.925 11-11 11S1 18.075 1 12 5.925 1 12 1ZM2.5 12a9.5 9.5 0 0 0 9.5 9.5 9.5 9.5 0 0 0 9.5-9.5A9.5 9.5 0 0 0 12 2.5 9.5 9.5 0 0 0 2.5 12Zm9.5 2a2 2 0 1 1-.001-3.999A2 2 0 0 1 12 14Z"></path>`
	iconComment = `<path d="M1.5 4.25c0-.966.784-1.75 1.75-1.75h17.5c.966 0 1.75.784 1.75 1.75v12.5a1.75 1.75 0 0 1-1.75 1.75h-9.69l-3.573 3.573A1.458 1.458 0 0 1 5 21.043V18.5H3.25a1.75 1.75 0 0 1-1.75-1.75ZM3.25 4a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h2.5a.75.75 0 0 1 .75.75v3.19l3.72-3.72a.749.749 0 0 1 .53-.22h10a.25.25 0 0 0 .25-.25V4.25a.25.25 0 0 0-.25-.25Z"></path>`
	iconProject = `<path d="M3 2.75A2.75 2.75 0 0 1 5.75 0h14.5a.75.75 0 0 1 .75.75v20.5a.75.75 0 0 1-.75.75h-6a.75.75 0 0 1 0-1.5h5.25v-4H6A1.5 1.5 0 0 0 4.5 18v.75c0 .716.43 1.334 1.05 1.605a.75.75 0 0 1-.6 1.374A3.251 3.251 0 0 1 3 18.75ZM19.5 1.5H5.75c-.69 0-1.25.56-1.25 1.25v12.651A2.989 2.989 0 0 1 6 15h13.5Z"></path><path d="M7 18.25a.25.25 0 0 1 .25-.25h5a.25.25 0 0 1 .25.25v5.01a.25.25 0 0 1-.397.201l-2.206-1.604a.25.25 0 0 0-.294 0L7.397 23.46a.25.25 0 0 1-.397-.2v-5.01Z"></path>`
	iconStar    = `<path d="M12 .25a.75.75 0 0 1 .673.418l3.058 6.197 6.839.994a.75.75 0 0 1 .415 1.279l-4.948 4.823 1.168 6.811a.751.751 0 0 1-1.088.791L12 18.347l-6.117 3.216a.75.75 0 0 1-1.088-.79l1.168-6.812-4.948-4.823a.75.75 0 0 1 .416-1.28l6.838-.993L11.328.668A.75.75 0 0 1 12 .25Zm0 2.445L9.44 7.882a.75.75 0 0 1-.565.41l-5.725.832 4.143 4.038a.748.748 0 0 1 .215.664l-.978 5.702 5.121-2.692a.75.75 0 0 1 .698 0l5.12 2.692-.977-5.702a.748.748 0 0 1 .215-.664l4.143-4.038-5.725-.831a.75.75 0 0 1-.565-.41L12 2.694Z"></path>`
)

// ---- Visual primitives ----
func renderSectionHeader(title string) string {
	return fmt.Sprintf(
		`<text x="0" y="20" font-family="system-ui, -apple-system, sans-serif" font-size="14" font-weight="700" fill="#9ca3af" letter-spacing="1">%s</text>`,
		html.EscapeString(title),
	)
}

func renderEmptyState(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf(
		`<text x="0" y="50" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">%s</text>`,
		html.EscapeString(message),
	)
}

func renderOwnedBadges(badge BadgeVM, cardWidth int) string {
	return fmt.Sprintf(`
	<g transform="translate(%d, 10.5)">
		%s
		<text x="32" y="16.5"
			font-family="system-ui, -apple-system, sans-serif"
			font-size="14"
			font-weight="600"
			fill="#22c55e">%s</text>
	</g>`,
		cardWidth-80,
		renderSmallIconBox(badge.Icon),
		badge.Count,
	)
}

func renderExternalBadges(badges []BadgeVM, cardWidth int) string {
	var sb strings.Builder
	currentX := -50

	for _, b := range badges {
		sb.WriteString(fmt.Sprintf(`
		<a xlink:href="%s" target="_blank">
			<g transform="translate(%d, 0)">
				%s
				<text x="28" y="16.5"
					font-family="system-ui, -apple-system, sans-serif"
					font-size="11"
					font-weight="700"
					fill="#22c55e">%s</text>
			</g>
		</a>`,
			html.EscapeString(b.Link),
			currentX,
			renderSmallIconBox(b.Icon),
			b.Count,
		))
		currentX -= 50
	}

	return fmt.Sprintf(`<g transform="translate(%d, 10.5)">%s</g>`, cardWidth-5, sb.String())
}
