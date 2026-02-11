package domain

import (
	"time"
)

type SemanticEventType string

const (
	SemanticEventPrOpened          SemanticEventType = "PR_OPENED"
	SemanticEventPrReview          SemanticEventType = "PR_REVIEW"         // Formal reviews only
	SemanticEventPrReviewComment   SemanticEventType = "PR_REVIEW_COMMENT" // Inline comments
	SemanticEventIssueOpened       SemanticEventType = "ISSUE_OPENED"
	SemanticEventIssueComment      SemanticEventType = "ISSUE_COMMENT"
	SemanticEventDiscussionOpened  SemanticEventType = "DISCUSSION_OPENED"
	SemanticEventDiscussionComment SemanticEventType = "DISCUSSION_COMMENT"
	SemanticEventCommit            SemanticEventType = "COMMIT"
)

type SemanticEvent struct {
	ID             string            `json:"id"`
	Type           SemanticEventType `json:"type"`
	Repo           string            `json:"repo"`
	AvatarURL      string            `json:"avatar_url"`
	URL            string            `json:"url"`
	Title          string            `json:"title,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	BaseScore      float64           `json:"base_score"`
	PopularityRaw  float64           `json:"popularity_raw"`
	Merged         bool              `json:"merged"`
	ReactionsCount int               `json:"reactions_count"`
}

// StatsView represents raw activity counts (unweighted).
type StatsView struct {
	PRsOpened               int
	PRReviews               int
	PRReviewComments        int
	IssuesOpened            int
	IssueComments           int
	DiscussionsOpened       int
	DiscussionComments      int
	TotalCommits            int
	ProjectsOwned           int
	StarsEarned             int
	TotalReposContributedTo int
}
