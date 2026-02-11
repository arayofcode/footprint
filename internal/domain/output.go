package domain

import (
	"time"
)

// RepoContribution represents a finalized, weighted output for an external repository.
type RepoContribution struct {
	Repo               string
	RepoURL            string
	AvatarURL          string
	Score              float64 // Final weighted score
	BaseScore          float64 // Sum of per-event base scores
	PopularityRaw      float64 // Peak popularity raw
	PRsOpened          int
	PRReviews          int
	PRReviewComments   int
	IssuesOpened       int
	IssueComments      int
	DiscussionsOpened  int
	DiscussionComments int
	Commits            int
	Events             []Contribution // Finalized output contributions
}

// FinalizedContributionType is an output-domain specific event type.
type FinalizedContributionType string

const (
	ContributionPR                FinalizedContributionType = "PR"
	ContributionPRReview          FinalizedContributionType = "PR_REVIEW"
	ContributionPRReviewComment   FinalizedContributionType = "PR_REVIEW_COMMENT"
	ContributionIssue             FinalizedContributionType = "ISSUE"
	ContributionIssueComment      FinalizedContributionType = "ISSUE_COMMENT"
	ContributionDiscussion        FinalizedContributionType = "DISCUSSION"
	ContributionDiscussionComment FinalizedContributionType = "DISCUSSION_COMMENT"
	ContributionCommit            FinalizedContributionType = "COMMIT"
	ContributionUnknown           FinalizedContributionType = "UNKNOWN"
)

// Contribution represents a finalized, renderable external activity.
type Contribution struct {
	Type           FinalizedContributionType
	Repo           string
	Title          string
	URL            string
	CreatedAt      time.Time
	ReactionsCount int
	Merged         bool
}

// MapSemanticToOutputEventType converts semantic internal types to output-safe types.
func MapSemanticToOutputEventType(st SemanticEventType) FinalizedContributionType {
	switch st {
	case SemanticEventPrOpened:
		return ContributionPR
	case SemanticEventPrReview:
		return ContributionPRReview
	case SemanticEventPrReviewComment:
		return ContributionPRReviewComment
	case SemanticEventIssueOpened:
		return ContributionIssue
	case SemanticEventIssueComment:
		return ContributionIssueComment
	case SemanticEventDiscussionOpened:
		return ContributionDiscussion
	case SemanticEventDiscussionComment:
		return ContributionDiscussionComment
	case SemanticEventCommit:
		return ContributionCommit
	default:
		return ContributionUnknown
	}
}

// MapEventsToContributions is an adapter that projects semantic events to output models.
func MapEventsToContributions(events []SemanticEvent) []Contribution {
	contribs := make([]Contribution, len(events))
	for i, e := range events {
		contribs[i] = Contribution{
			Type:           MapSemanticToOutputEventType(e.Type),
			Repo:           e.Repo,
			Title:          e.Title,
			URL:            e.URL,
			CreatedAt:      e.CreatedAt,
			ReactionsCount: e.ReactionsCount,
			Merged:         e.Merged,
		}
	}
	return contribs
}

// OwnedProjectImpact represents a finalized, weighted output for an owned repository.
type OwnedProjectImpact struct {
	Repo          string
	URL           string
	AvatarURL     string
	Stars         int
	Forks         int
	BaseScore     float64
	PopularityRaw float64
	Score         float64 // Final weighted score
}
