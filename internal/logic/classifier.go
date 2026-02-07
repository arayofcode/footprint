package logic

import (
	"strings"

	"github.com/arayofcode/footprint/internal/domain"
)

func Classify(e domain.ContributionEvent) domain.SemanticEvent {
	semanticType := domain.SemanticEventIssueComment

	switch e.Type {
	case domain.ContributionTypePR:
		semanticType = domain.SemanticEventPrOpened
	case domain.ContributionTypePullRequestReview, domain.ContributionTypePullRequestReviewComment:
		semanticType = domain.SemanticEventPrFeedback
	case domain.ContributionTypeIssue:
		semanticType = domain.SemanticEventIssueOpened
	case domain.ContributionTypeIssueComment, domain.ContributionTypePRComment:
		if strings.Contains(e.URL, "/pull/") {
			semanticType = domain.SemanticEventPrFeedback
		} else {
			semanticType = domain.SemanticEventIssueComment
		}
	case domain.ContributionTypeDiscussion, domain.ContributionTypeDiscussionComment:
		// Map discussions to issue comments for now, or new type later
		semanticType = domain.SemanticEventIssueComment
	}

	return domain.SemanticEvent{
		ID:        e.ID,
		Type:      semanticType,
		Repo:      e.Repo,
		AvatarURL: e.RepoOwnerAvatarURL,
		URL:       e.URL,
		Title:     e.Title,
		CreatedAt: e.CreatedAt,
		Score:     e.Score,
		Merged:    e.Merged,
	}
}

func MapClassify(events []domain.ContributionEvent) []domain.SemanticEvent {
	semantic := make([]domain.SemanticEvent, len(events))
	for i, e := range events {
		semantic[i] = Classify(e)
	}
	return semantic
}
