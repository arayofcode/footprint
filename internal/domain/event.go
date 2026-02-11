package domain

import (
	"fmt"
	"math"
	"time"
)

type ContributionType string

const (
	ContributionTypePR                ContributionType = "PR"
	ContributionTypePRComment         ContributionType = "PR_COMMENT"
	ContributionTypeIssue             ContributionType = "ISSUE"
	ContributionTypeIssueComment      ContributionType = "ISSUE_COMMENT"
	ContributionTypeReview            ContributionType = "REVIEW" // Was PR_FEEDBACK
	ContributionTypeReviewComment     ContributionType = "REVIEW_COMMENT"
	ContributionTypeDiscussion        ContributionType = "DISCUSSION"
	ContributionTypeDiscussionComment ContributionType = "DISCUSSION_COMMENT"
)

type ContributionEvent struct {
	ID                 string           `json:"id"`
	Type               ContributionType `json:"type"`
	Repo               string           `json:"repo"`
	RepoOwnerAvatarURL string           `json:"repo_owner_avatar_url,omitempty"`
	URL                string           `json:"url"`
	Title              string           `json:"title,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	Stars              int              `json:"stars,omitempty"`
	Forks              int              `json:"forks,omitempty"`
	Merged             bool             `json:"is_merged,omitempty"`
	Answer             bool             `json:"is_answer,omitempty"`
	Snippet            string           `json:"snippet,omitempty"`
	ReactionsCount     int              `json:"reactions_count,omitempty"`
	BaseScore          float64          `json:"base_score,omitempty"`
	PopularityRaw      float64          `json:"popularity_raw,omitempty"`
}

// Deprecated: Use StatsView instead.
type UserStats struct {
	TotalCommits       int
	TotalPRs           int
	TotalIssues        int
	TotalReviews       int
	TotalIssueComments int
	TotalReposCount    int
	TotalStarsEarned   int
}

type OwnedProject struct {
	Repo      string `json:"repo"`
	URL       string `json:"url"`
	AvatarURL string `json:"avatar_url"`
	Stars     int    `json:"stars"`
	Forks     int    `json:"forks"`
}

type EnrichedProject struct {
	OwnedProject
	BaseScore     float64
	PopularityRaw float64
}

func (e ContributionEvent) StableID() string {
	if e.ID != "" {
		return e.ID
	}
	if e.Repo != "" && e.URL != "" {
		return fmt.Sprintf("%s#%s", e.Repo, e.URL)
	}
	return ""
}

func (e ContributionEvent) PopularityMultiplier() float64 {
	return popularityMultiplier(e.Stars, e.Forks)
}

func (p OwnedProject) PopularityMultiplier() float64 {
	return popularityMultiplier(p.Stars, p.Forks)
}

func popularityMultiplier(stars, forks int) float64 {
	return 1 + math.Log10(1+float64(stars)+2*float64(forks))
}
