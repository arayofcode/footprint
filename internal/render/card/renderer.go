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
	ownedOSSCount := len(projects)

	externalRepoMap := make(map[string]*externalRepoStat)
	mergedPRCount := 0

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
			if e.Merged {
				mergedPRCount++
			}
		}
	}

	externalRepoCount := len(externalRepoMap)

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
	currentY := 225
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
<svg width="720" height="%d" viewBox="0 0 720 %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
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
  <rect width="720" height="%d" rx="20" fill="#1a1a1a" />
  
  <!-- Header section -->
  <g>
    <image href="%s" x="42" y="32" width="36" height="36" clip-path="url(#circle-clip)" />
    <circle cx="60" cy="50" r="18" fill="none" stroke="#6b7280" stroke-width="2"/>
    <text x="95" y="58" font-family="system-ui, -apple-system, sans-serif" font-size="26" font-weight="600" fill="white">GitHub Footprint</text>
  </g>
  
  <!-- Stat boxes -->
  <g transform="translate(40, 100)">
    <g>
      <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
      <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
      <path d="M 30 15 L 33 25 L 44 25 L 35 32 L 38 43 L 30 36 L 22 43 L 25 32 L 16 25 L 27 25 Z" fill="none" stroke="#22c55e" stroke-width="2" stroke-linejoin="round"/>
      <text x="75" y="45" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
      <text x="0" y="85" font-family="system-ui, -apple-system, sans-serif" font-size="16" fill="#9ca3af">Owned OSS Projects</text>
    </g>
    
    <g transform="translate(230, 0)">
      <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
      <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
      <circle cx="22" cy="20" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
      <circle cx="38" cy="20" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
      <circle cx="30" cy="45" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
      <path d="M 22 24 L 22 30 Q 22 35 30 35 L 30 41" fill="none" stroke="#22c55e" stroke-width="2"/>
      <path d="M 38 24 L 38 30 Q 38 35 30 35" fill="none" stroke="#22c55e" stroke-width="2"/>
      <text x="75" y="45" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
      <text x="0" y="85" font-family="system-ui, -apple-system, sans-serif" font-size="16" fill="#9ca3af">External Repos</text>
    </g>
    
    <g transform="translate(460, 0)">
      <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
      <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
      <circle cx="25" cy="20" r="4" fill="#22c55e"/>
      <line x1="25" y1="24" x2="25" y2="45" stroke="#22c55e" stroke-width="2"/>
      <circle cx="25" cy="48" r="3" fill="none" stroke="#22c55e" stroke-width="2"/>
      <path d="M 36 28 L 42 34 L 52 22" fill="none" stroke="#22c55e" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <text x="75" y="45" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
      <text x="0" y="85" font-family="system-ui, -apple-system, sans-serif" font-size="16" fill="#9ca3af">PRs merged</text>
    </g>
  </g>
  
  %s
  %s
  
  <g transform="translate(40, %d)">
    <text x="0" y="0" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Range: All-time</text>
    <a xlink:href="https://github.com/arayofcode/footprint" target="_blank">
      <text x="360" y="0" text-anchor="middle" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#22c55e" style="cursor: pointer;">Generated by Footprint</text>
    </a>
    <text x="640" y="0" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Last Updated %s</text>
  </g>
</svg>
`,
		totalHeight,
		totalHeight,
		totalHeight,
		userAvatarBase64,
		ownedOSSCount,
		externalRepoCount,
		mergedPRCount,
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
        <text x="640" y="18" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="18" fill="#22c55e">%s ★</text>
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
        <text x="640" y="18" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="18" font-weight="600" fill="#22c55e">%d PRs</text>
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
	defer resp.Body.Close()
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
