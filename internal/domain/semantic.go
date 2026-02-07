package domain

import (
	"time"
)

type SemanticEventType string

const (
	SemanticEventPrOpened     SemanticEventType = "PR_OPENED"
	SemanticEventPrFeedback   SemanticEventType = "PR_FEEDBACK"
	SemanticEventIssueOpened  SemanticEventType = "ISSUE_OPENED"
	SemanticEventIssueComment SemanticEventType = "ISSUE_COMMENT"
)

type SemanticEvent struct {
	ID        string            `json:"id"`
	Type      SemanticEventType `json:"type"`
	Repo      string            `json:"repo"`
	AvatarURL string            `json:"avatar_url"`
	URL       string            `json:"url"`
	Title     string            `json:"title,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Score     float64           `json:"score"`
	Merged    bool              `json:"merged"`
}

type StatsView struct {
	PRsOpened               int
	PRFeedback              int
	IssuesOpened            int
	IssueComments           int
	ProjectsOwned           int
	StarsEarned             int
	TotalReposContributedTo int
}

type RepoContribution struct {
	Repo          string
	RepoURL       string
	AvatarURL     string
	Score         float64
	PRsOpened     int
	PRFeedback    int
	IssuesOpened  int
	IssueComments int
}
