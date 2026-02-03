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
    <text x="95" y="58" font-family="system-ui, -apple-system, sans-serif" font-size="26" font-weight="600" fill="white">GitHub Footprint</text>
  </g>
  

  
%s
  %s
  %s
  
  <g transform="translate(40, %d)">
    <text x="0" y="0" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Range: All-time</text>
    <a xlink:href="https://github.com/arayofcode/footprint" target="_blank">
      <text x="250" y="0" text-anchor="middle" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#22c55e" style="cursor: pointer;">Generated by Footprint</text>
    </a>
    <text x="420" y="0" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Last Updated %s</text>
  </g>
</svg>
`,
		totalHeight,
		totalHeight,
		totalHeight,
		userAvatarBase64,

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

func renderStatBox(x, y int, label, value, iconPath, color string) string {
	return fmt.Sprintf(`
    <g transform="translate(%d, %d)">
      <rect x="-5" y="-5" width="70" height="70" rx="14" fill="%s" opacity="0.15" filter="url(#glow)"/>
      <rect width="60" height="60" rx="12" fill="#1f2937" stroke="%s" stroke-width="1.5" opacity="0.9"/>
      <g transform="translate(18, 18)">
        <path d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </g>
      <text x="75" y="45" font-family="system-ui, -apple-system, sans-serif" font-size="48" font-weight="700" fill="white">%s</text>
      <text x="75" y="68" font-family="system-ui, -apple-system, sans-serif" font-size="12" font-weight="600" fill="#9ca3af" letter-spacing="0.5">%s</text>
    </g>`, x, y, color, color, iconPath, color, value, strings.ToUpper(label))
}

const (
	iconPR      = "M6 3v12M18 9v12M6 21a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM18 9a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM6 9a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z m12 12a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z"                                 // Git branch-ish
	iconReview  = "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"                          // Clipboard-ish
	iconIssue   = "M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2m0 18a8 8 0 1 1 8-8 8 8 0 0 1-8 8m-1-5h2m-2-4h2m-4 4h2M9 9h2m-2 4h2m4-4h2m-2 4h2"                                                                 // Bug/Circle-ish
	iconComment = "M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" // Message bubble
	iconProject = "M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z M3.27 6.96L12 12.01l8.73-5.05 M12 22.08V12"                     // Box/Package
	iconStar    = "M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"                                                                                             // Star
)
