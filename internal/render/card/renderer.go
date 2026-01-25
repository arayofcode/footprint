package card

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type Renderer struct{}

type repoStat struct {
	name       string
	prCount    int
	issueCount int
	otherCount int
	score      float64
	avatarURL  string
}

func (Renderer) RenderCard(ctx context.Context, user domain.User, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	_ = ctx

	// 1. Calculate Stats
	mergedPRs := 0
	uniqueRepos := make(map[string]bool)
	for _, e := range events {
		if e.Type == domain.ContributionTypePR && e.Merged {
			mergedPRs++
		}
		uniqueRepos[e.Repo] = true
	}

	// 2. Rank Top Repos by Impact Score
	repoMap := make(map[string]*repoStat)
	for _, e := range events {
		if _, ok := repoMap[e.Repo]; !ok {
			repoMap[e.Repo] = &repoStat{
				name:      e.Repo,
				avatarURL: e.RepoOwnerAvatarURL,
			}
		}
		repoMap[e.Repo].score += e.Score

		switch e.Type {
		case domain.ContributionTypePR:
			repoMap[e.Repo].prCount++
		case domain.ContributionTypeIssue:
			repoMap[e.Repo].issueCount++
		default:
			repoMap[e.Repo].otherCount++
		}
	}

	var rankedRepos []*repoStat
	for _, r := range repoMap {
		rankedRepos = append(rankedRepos, r)
	}
	sort.Slice(rankedRepos, func(i, j int) bool {
		if rankedRepos[i].score == rankedRepos[j].score {
			return rankedRepos[i].name < rankedRepos[j].name
		}
		return rankedRepos[i].score > rankedRepos[j].score
	})

	// Only take top 3
	topRepos := rankedRepos
	if len(topRepos) > 3 {
		topRepos = topRepos[:3]
	}

	// 3. Fetch and Base64 encode User Avatar for embedding
	userAvatarBase64 := fetchAsDataURL(user.AvatarURL)

	// 4. Generate SVG using the target template
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="720" height="520" viewBox="0 0 720 520" fill="none" xmlns="http://www.w3.org/2000/svg">
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
       <rect width="32" height="32" rx="6" />
    </clipPath>
  </defs>
  
  <!-- Background -->
  <rect width="720" height="520" rx="20" fill="#1a1a1a" />
  
  <!-- Header section -->
  <!-- User icon -->
  <g>
    <image href="%s" x="42" y="32" width="36" height="36" clip-path="url(#circle-clip)" />
    <circle cx="60" cy="50" r="18" fill="none" stroke="#6b7280" stroke-width="2"/>
  </g>
  
  <text x="95" y="58" font-family="system-ui, -apple-system, sans-serif" font-size="26" font-weight="600" fill="white">GitHub Contributions</text>
  
  <!-- Stat boxes -->
  <g transform="translate(40, 100)">
    <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
    <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
    <circle cx="25" cy="20" r="4" fill="#22c55e"/>
    <line x1="25" y1="24" x2="25" y2="45" stroke="#22c55e" stroke-width="2"/>
    <circle cx="25" cy="48" r="3" fill="none" stroke="#22c55e" stroke-width="2"/>
    <path d="M 36 28 L 42 34 L 52 22" fill="none" stroke="#22c55e" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
  </g>
  <text x="120" y="145" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
  <text x="40" y="195" font-family="system-ui, -apple-system, sans-serif" font-size="18" fill="#9ca3af">Merged PRs</text>
  
  <g transform="translate(270, 100)">
    <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
    <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
    <path d="M 30 15 L 33 25 L 44 25 L 35 32 L 38 43 L 30 36 L 22 43 L 25 32 L 16 25 L 27 25 Z" fill="none" stroke="#22c55e" stroke-width="2" stroke-linejoin="round"/>
  </g>
  <text x="350" y="145" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
  <text x="260" y="195" font-family="system-ui, -apple-system, sans-serif" font-size="18" fill="#9ca3af">Popular Projects</text>
  
  <g transform="translate(500, 100)">
    <rect x="-5" y="-5" width="70" height="70" rx="14" fill="#22c55e" opacity="0.15" filter="url(#glow)"/>
    <rect width="60" height="60" rx="12" fill="#1f2937" stroke="#22c55e" stroke-width="1.5" opacity="0.9"/>
    <circle cx="22" cy="20" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
    <circle cx="38" cy="20" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
    <circle cx="30" cy="45" r="4" fill="none" stroke="#22c55e" stroke-width="2"/>
    <path d="M 22 24 L 22 30 Q 22 35 30 35 L 30 41" fill="none" stroke="#22c55e" stroke-width="2"/>
    <path d="M 38 24 L 38 30 Q 38 35 30 35" fill="none" stroke="#22c55e" stroke-width="2"/>
  </g>
  <text x="580" y="145" font-family="system-ui, -apple-system, sans-serif" font-size="52" font-weight="700" fill="white">%d</text>
  <text x="480" y="195" font-family="system-ui, -apple-system, sans-serif" font-size="18" fill="#9ca3af">Repos Contributed</text>
  
  <text x="40" y="270" font-family="system-ui, -apple-system, sans-serif" font-size="28" font-weight="600" fill="white">Top Impact</text>
  
  %s
  
  <text x="680" y="495" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="12" fill="#6b7280">Updated %s</text>
</svg>
`,
		userAvatarBase64,
		mergedPRs,
		len(projects),
		len(uniqueRepos),
		formatTopRepos(topRepos),
		generatedAt.Format("02 Jan 2006"),
	)

	return []byte(svg), nil
}

func formatTopRepos(repos []*repoStat) string {
	if len(repos) == 0 {
		return `<text x="40" y="310" font-family="system-ui, -apple-system, sans-serif" font-size="22" fill="#9ca3af">No public contributions discovered yet.</text>`
	}

	var s string
	for i, r := range repos {
		y := 300 + (i * 65)
		repoAvatarBase64 := fetchAsDataURL(r.avatarURL)

		var parts []string
		if r.prCount > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", r.prCount, pluralize(r.prCount, "PR", "PRs")))
		}
		if r.issueCount > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", r.issueCount, pluralize(r.issueCount, "Issue", "Issues")))
		}
		if r.otherCount > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", r.otherCount, pluralize(r.otherCount, "Act", "Acts")))
		}

		summary := ""
		if len(parts) > 0 {
			summary = parts[0]
			for j := 1; j < len(parts); j++ {
				summary += " · " + parts[j]
			}
		}

		s += fmt.Sprintf(`
  <g transform="translate(40, %d)">
    <circle cx="8" cy="25" r="4" fill="#22c55e"/>
    <g transform="translate(30, 10)">
       <rect width="32" height="32" rx="6" fill="#1f2937" />
       <image href="%s" width="32" height="32" clip-path="url(#repo-clip)" />
    </g>
    <text x="80" y="33" font-family="system-ui, -apple-system, sans-serif" font-size="22" font-weight="600" fill="white">%s</text>
    <text x="500" y="33" text-anchor="end" font-family="system-ui, -apple-system, sans-serif" font-size="22" fill="#6b7280">•</text>
    <text x="520" y="33" font-family="system-ui, -apple-system, sans-serif" font-size="22" font-style="italic" fill="#22c55e">%s</text>
  </g>`,
			y, repoAvatarBase64, truncate(r.name, 30), summary)
	}
	return s
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
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
		return html.EscapeString(url) // Fallback to URL if fetch fails, escaped for SVG
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return html.EscapeString(url) // Fallback to URL if status is not OK, escaped for SVG
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return html.EscapeString(url) // Fallback to URL if read fails, escaped for SVG
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png" // Guessing PNG as fallback
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
}
