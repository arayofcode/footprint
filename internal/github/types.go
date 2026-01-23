package github

import "time"

type ContributionType string

const (
	ContributionTypePR                ContributionType = "PR"
	ContributionTypeIssue             ContributionType = "ISSUE"
	ContributionTypeIssueComment      ContributionType = "ISSUE_COMMENT"
	ContributionTypeReview            ContributionType = "REVIEW"
	ContributionTypeReviewComment     ContributionType = "REVIEW_COMMENT"
	ContributionTypeDiscussion        ContributionType = "DISCUSSION"
	ContributionTypeDiscussionComment ContributionType = "DISCUSSION_COMMENT"
)

type ContributionEvent struct {
	ID             string           `json:"id"`
	Type           ContributionType `json:"type"`
	Repo           string           `json:"repo"`
	URL            string           `json:"url"`
	Title          string           `json:"title,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	Snippet        string           `json:"snippet,omitempty"`
	ReactionsCount int              `json:"reactions_count,omitempty"`
	Score          float64          `json:"score,omitempty"`
	Merged         bool             `json:"is_merged,omitempty"`
	Answer         bool             `json:"is_answer,omitempty"`
}
