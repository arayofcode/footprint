package scoring

import (
	"fmt"

	"github.com/arayofcode/footprint/internal/domain"
)

const (
	DefaultClamp   = 10.0
	MergedPRBonus  = 1.5
	OwnershipScore = 2500.0
)

var baseContributionScores = map[domain.ContributionType]float64{
	domain.ContributionTypePR:                10.0,
	domain.ContributionTypeIssue:             5.0,
	domain.ContributionTypeIssueComment:      2.0,
	domain.ContributionTypeReview:            3.0,
	domain.ContributionTypeReviewComment:     2.0,
	domain.ContributionTypeDiscussion:        2.0,
	domain.ContributionTypeDiscussionComment: 2.0,
}

type Calculator struct {
	Clamp float64
}

func NewCalculator(clamp float64) *Calculator {
	if clamp <= 0 {
		clamp = DefaultClamp
	}
	return &Calculator{Clamp: clamp}
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

func (c *Calculator) ScoreOwnedProject(project domain.OwnedProject) domain.OwnedProject {
	// Weighted ownership impact based on project popularity
	project.Score = OwnershipScore * project.PopularityMultiplier(c.Clamp)
	return project
}

func baseScore(event domain.ContributionEvent) float64 {
	if score, ok := baseContributionScores[event.Type]; ok {
		return score
	}
	fmt.Printf("Score not found for type: %s", event.Type)
	return 0
}
