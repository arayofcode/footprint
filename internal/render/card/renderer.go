package card

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

type externalRepoStat struct {
	name      string
	url       string
	avatarURL string
	score     float64
	prCount   int
}

func (Renderer) RenderCard(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	// 1. Calculate KPIs
	externalRepoMap := make(map[string]*externalRepoStat)

	for _, e := range events {
		// Only consider truly external contributions
		parts := strings.Split(e.Repo, "/")
		if len(parts) < 2 || parts[0] == user.Username {
			continue
		}

		if _, ok := externalRepoMap[e.Repo]; !ok {
			repoURL := fmt.Sprintf("https://github.com/%s", e.Repo)
			externalRepoMap[e.Repo] = &externalRepoStat{
				name:      e.Repo,
				url:       repoURL,
				avatarURL: e.RepoOwnerAvatarURL,
			}
		}
		externalRepoMap[e.Repo].score += e.Score

		if e.Type == domain.ContributionTypePR {
			externalRepoMap[e.Repo].prCount++
		}
	}

	// Calculate total stars
	totalStars := 0
	for _, p := range projects {
		totalStars += p.Stars
	}

	// 6 Stats Grid
	// Row 1
	stat1 := renderStatBox(0, 0, "PRs Opened", formatCount(stats.TotalPRs), iconPR, "#22c55e")
	stat2 := renderStatBox(230, 0, "PR Reviews", formatCount(stats.TotalReviews), iconReview, "#22c55e")

	// Row 2
	stat3 := renderStatBox(0, 100, "Issues Opened", formatCount(stats.TotalIssues), iconIssue, "#22c55e")
	stat4 := renderStatBox(230, 100, "Issue Comments", formatCount(stats.TotalIssueComments), iconComment, "#22c55e")

	// Row 3
	stat5 := renderStatBox(0, 200, "Projects Owned", formatCount(len(projects)), iconProject, "#22c55e")
	stat6 := renderStatBox(230, 200, "Stars Earned", formatLargeNum(totalStars), iconStar, "#22c55e")

	statSection := fmt.Sprintf(`
  <g transform="translate(40, 100)">
    %s
    %s
    %s
    %s
    %s
    %s
  </g>`, stat1, stat2, stat3, stat4, stat5, stat6)

	// 2. Prepare Sections Data

	// Owned Projects (Top 3)
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Stars != projects[j].Stars {
			return projects[i].Stars > projects[j].Stars
		}
		if projects[i].Forks != projects[j].Forks {
			return projects[i].Forks > projects[j].Forks
		}
		return projects[i].Repo < projects[j].Repo
	})
	topOwned := projects
	if len(topOwned) > 3 {
		topOwned = topOwned[:3]
	}

	// External Contributions (Top 3)
	var rankedExternal []*externalRepoStat
	for _, r := range externalRepoMap {
		rankedExternal = append(rankedExternal, r)
	}
	sort.Slice(rankedExternal, func(i, j int) bool {
		if rankedExternal[i].score != rankedExternal[j].score {
			return rankedExternal[i].score > rankedExternal[j].score
		}
		return rankedExternal[i].name < rankedExternal[j].name
	})
	topExternal := rankedExternal
	if len(topExternal) > 3 {
		topExternal = topExternal[:3]
	}

	// 3. Dynamic Section Positioning
	// Stat grid now has 3 rows (0, 100, 200). Ends at 200+70 = 270.
	// Add gap of 40 -> 310 relative to start (100) -> 410 absolute.
	currentY := 410
	ownedSection := ""
	if len(topOwned) > 0 {
		ownedSection = fmt.Sprintf(`
  <g transform="translate(40, %d)">
    <text x="0" y="20" font-family="system-ui, -apple-system, sans-serif" font-size="24" font-weight="600" fill="white">Owned Projects</text>
    %s
  </g>`, currentY, formatOwned(topOwned))
		currentY += 45 + (len(topOwned) * 45)
	}

	externalSection := ""
	if len(topExternal) > 0 {
		if ownedSection != "" {
			currentY += 20 // Gap between sections
		}
		externalSection = fmt.Sprintf(`
  <g transform="translate(40, %d)">
    <text x="0" y="20" font-family="system-ui, -apple-system, sans-serif" font-size="24" font-weight="600" fill="white">Top Repositories</text>
    %s
  </g>`, currentY, formatExternal(topExternal, user.Username))
		currentY += 45 + (len(topExternal) * 45)
	}

	totalHeight := currentY + 60 // Add space for footer and padding
	footerY := totalHeight - 25

	// 4. Fetch and Base64 encode User Avatar
	userAvatarBase64 := fetchAsDataURL(user.AvatarURL)

	// 5. Generate SVG
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="500" height="%d" viewBox="0 0 500 %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <defs>
    <filter id="glow" x="-50%%" y="-50%%" width="200%%" height="200%%">
      <feGaussianBlur stdDeviation="8" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    <clipPath id="circle-clip">
      <circle cx="60" cy="50" r="18" />
    </clipPath>
    <clipPath id="repo-clip">
       <rect width="24" height="24" rx="4" />
    </clipPath>
  </defs>
  
  <!-- Background -->
  <rect width="500" height="%d" rx="20" fill="#1a1a1a" />
  
  <!-- Header section -->
  <g>
    <image href="%s" x="42" y="32" width="36" height="36" clip-path="url(#circle-clip)" />
    <circle cx="60" cy="50" r="18" fill="none" stroke="#6b7280" stroke-width="2"/>
    <text x="95" y="58" font-family="system-ui, -apple-system, sans-serif" font-size="26" font-weight="600" fill="white">%s's footprint</text>
  </g>
  

  
%s
  %s
  %s
  
  <g transform="translate(40, %d)">
    <text x="0" y="0" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Range: All-time</text>
    <text x="420" y="0" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Last Updated %s</text>
  </g>
</svg>
`,
		totalHeight,
		totalHeight,
		totalHeight,
		userAvatarBase64,
		user.Username,
		statSection,
		ownedSection,
		externalSection,
		footerY,
		generatedAt.Format("02 Jan 2006"),
	)

	return []byte(svg), nil
}

func formatOwned(projects []domain.OwnedProject) string {
	var s string
	for i, p := range projects {
		y := 45 + (i * 45)
		repoAvatar := fetchAsDataURL(p.AvatarURL)
		s += fmt.Sprintf(`
    <a xlink:href="%s" target="_blank" style="cursor: pointer;">
      <g transform="translate(0, %d)">
        <image href="%s" width="24" height="24" clip-path="url(#repo-clip)" x="0" y="0"/>
        <text x="35" y="18" font-family="system-ui, -apple-system, sans-serif" font-size="18" font-weight="500" fill="white">%s</text>
        <text x="420" y="18" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="18" fill="#22c55e">%s ★</text>
      </g>
    </a>`,
			html.EscapeString(p.URL), y, repoAvatar, truncate(p.Repo, 40), formatCount(p.Stars))
	}
	return s
}

func formatExternal(repos []*externalRepoStat, username string) string {
	var s string
	for i, r := range repos {
		y := 45 + (i * 45)
		repoAvatar := fetchAsDataURL(r.avatarURL)

		// Target user PRs in this repo
		prLink := fmt.Sprintf("https://github.com/%s/pulls?q=is%%3Apr+author%%3A%s", r.name, username)

		s += fmt.Sprintf(`
    <a xlink:href="%s" target="_blank" style="cursor: pointer;">
      <g transform="translate(0, %d)">
        <image href="%s" width="24" height="24" clip-path="url(#repo-clip)" x="0" y="0"/>
        <text x="35" y="18" font-family="system-ui, -apple-system, sans-serif" font-size="18" font-weight="500" fill="white">%s</text>
        <text x="420" y="18" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="#22c55e">%d PRs</text>
      </g>
    </a>`,
			html.EscapeString(prLink), y, repoAvatar, truncate(r.name, 40), r.prCount)
	}
	return s
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

func fetchAsDataURL(url string) string {
	if url == "" {
		return ""
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return html.EscapeString(url)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return html.EscapeString(url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return html.EscapeString(url)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
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
      <text x="75" y="40" font-family="system-ui, -apple-system, sans-serif" font-size="38" font-weight="700" fill="white">%s</text>
      <text x="75" y="57" font-family="system-ui, -apple-system, sans-serif" font-size="10" font-weight="600" fill="#9ca3af" letter-spacing="0.5">%s</text>
    </g>`, x, y, color, color, color, iconXML, value, strings.ToUpper(label))
}

const (
	iconPR      = `<path d="M16 19.25a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0Zm-14.5 0a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0Zm0-14.5a3.25 3.25 0 1 1 6.5 0 3.25 3.25 0 0 1-6.5 0ZM4.75 3a1.75 1.75 0 1 0 .001 3.501A1.75 1.75 0 0 0 4.75 3Zm0 14.5a1.75 1.75 0 1 0 .001 3.501A1.75 1.75 0 0 0 4.75 17.5Zm14.5 0a1.75 1.75 0 1 0 .001 3.501 1.75 1.75 0 0 0-.001-3.501Z"></path><path d="M13.405 1.72a.75.75 0 0 1 0 1.06L12.185 4h4.065A3.75 3.75 0 0 1 20 7.75v8.75a.75.75 0 0 1-1.5 0V7.75a2.25 2.25 0 0 0-2.25-2.25h-4.064l1.22 1.22a.75.75 0 0 1-1.061 1.06l-2.5-2.5a.75.75 0 0 1 0-1.06l2.5-2.5a.75.75 0 0 1 1.06 0ZM4.75 7.25A.75.75 0 0 1 5.5 8v8A.75.75 0 0 1 4 16V8a.75.75 0 0 1 .75-.75Z"></path>`
	iconReview  = `<path d="M10.3 6.74a.75.75 0 0 1-.04 1.06l-2.908 2.7 2.908 2.7a.75.75 0 1 1-1.02 1.1l-3.5-3.25a.75.75 0 0 1 0-1.1l3.5-3.25a.75.75 0 0 1 1.06.04Zm3.44 1.06a.75.75 0 1 1 1.02-1.1l3.5 3.25a.75.75 0 0 1 0 1.1l-3.5 3.25a.75.75 0 1 1-1.02-1.1l2.908-2.7-2.908-2.7Z"></path><path d="M1.5 4.25c0-.966.784-1.75 1.75-1.75h17.5c.966 0 1.75.784 1.75 1.75v12.5a1.75 1.75 0 0 1-1.75 1.75h-9.69l-3.573 3.573A1.458 1.458 0 0 1 5 21.043V18.5H3.25a1.75 1.75 0 0 1-1.75-1.75ZM3.25 4a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h2.5a.75.75 0 0 1 .75.75v3.19l3.72-3.72a.749.749 0 0 1 .53-.22h10a.25.25 0 0 0 .25-.25V4.25a.25.25 0 0 0-.25-.25Z"></path>`
	iconIssue   = `<path d="M12 1c6.075 0 11 4.925 11 11s-4.925 11-11 11S1 18.075 1 12 5.925 1 12 1ZM2.5 12a9.5 9.5 0 0 0 9.5 9.5 9.5 9.5 0 0 0 9.5-9.5A9.5 9.5 0 0 0 12 2.5 9.5 9.5 0 0 0 2.5 12Zm9.5 2a2 2 0 1 1-.001-3.999A2 2 0 0 1 12 14Z"></path>`
	iconComment = `<path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>`
	iconProject = `<path d="M3 2.75A2.75 2.75 0 0 1 5.75 0h14.5a.75.75 0 0 1 .75.75v20.5a.75.75 0 0 1-.75.75h-6a.75.75 0 0 1 0-1.5h5.25v-4H6A1.5 1.5 0 0 0 4.5 18v.75c0 .716.43 1.334 1.05 1.605a.75.75 0 0 1-.6 1.374A3.251 3.251 0 0 1 3 18.75ZM19.5 1.5H5.75c-.69 0-1.25.56-1.25 1.25v12.651A2.989 2.989 0 0 1 6 15h13.5Z"></path><path d="M7 18.25a.25.25 0 0 1 .25-.25h5a.25.25 0 0 1 .25.25v5.01a.25.25 0 0 1-.397.201l-2.206-1.604a.25.25 0 0 0-.294 0L7.397 23.46a.25.25 0 0 1-.397-.2v-5.01Z"></path>`
	iconStar    = `<path d="M12 .25a.75.75 0 0 1 .673.418l3.058 6.197 6.839.994a.75.75 0 0 1 .415 1.279l-4.948 4.823 1.168 6.811a.751.751 0 0 1-1.088.791L12 18.347l-6.117 3.216a.75.75 0 0 1-1.088-.79l1.168-6.812-4.948-4.823a.75.75 0 0 1 .416-1.28l6.838-.993L11.328.668A.75.75 0 0 1 12 .25Zm0 2.445L9.44 7.882a.75.75 0 0 1-.565.41l-5.725.832 4.143 4.038a.748.748 0 0 1 .215.664l-.978 5.702 5.121-2.692a.75.75 0 0 1 .698 0l5.12 2.692-.977-5.702a.748.748 0 0 1 .215-.664l4.143-4.038-5.725-.831a.75.75 0 0 1-.565-.41L12 2.694Z"></path>`
)
