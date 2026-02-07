package scoring

import (
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
)

var baseContributionScores = map[domain.ContributionType]float64{
	domain.ContributionTypePR:                10.0,
	domain.ContributionTypeIssue:             5.0,
	domain.ContributionTypeIssueComment:      2.0,
	domain.ContributionTypeReview:            3.0,
	domain.ContributionTypeReviewComment:     1.0,
	domain.ContributionTypeDiscussion:        2.0,
	domain.ContributionTypeDiscussionComment: 2.0,
}

func baseScore(event domain.ContributionEvent) float64 {
	if score, ok := baseContributionScores[event.Type]; ok {
		return score
	}
	fmt.Printf("Score not found for type: %s", event.Type)
	return 0
}

func (c *Calculator) ScoreContribution(event domain.ContributionEvent) domain.ContributionEvent {
	base := baseScore(event)
	multiplier := event.PopularityMultiplier(c.Clamp)

	if event.Type == domain.ContributionTypePR && event.Merged {
		base = base * MergedPRBonus
	}

	event.Score = base * multiplier
	return event
}
