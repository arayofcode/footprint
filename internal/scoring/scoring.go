package scoring

import (
	"sort"

	"github.com/arayofcode/footprint/internal/domain"
)

const (
	MergedPRBonus  = 1.5
	OwnershipScore = 2500.0
)

type Calculator struct {
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) ScoreBatch(events []domain.ContributionEvent) []domain.ContributionEvent {
	// Map to track counts: repo -> type -> count
	counts := make(map[string]map[domain.ContributionType]int)

	scored := make([]domain.ContributionEvent, len(events))
	copy(scored, events)

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].CreatedAt.Equal(scored[j].CreatedAt) {
			return scored[i].URL < scored[j].URL
		}
		return scored[i].CreatedAt.Before(scored[j].CreatedAt)
	})

	for i := range scored {
		repo := scored[i].Repo
		if _, ok := counts[repo]; !ok {
			counts[repo] = make(map[domain.ContributionType]int)
		}

		// Increment count for this type in this repo
		count := counts[repo][scored[i].Type]
		counts[repo][scored[i].Type]++

		// Standard score calculation
		scored[i] = c.ScoreContribution(scored[i])

		if isDecayable(scored[i].Type) {
			decay := 1.0 / (1.0 + 0.5*float64(count)) // 1, 0.66, 0.5, 0.4...
			scored[i].BaseScore *= decay
		}
	}
	return scored
}

func isDecayable(t domain.ContributionType) bool {
	return t == domain.ContributionTypeIssueComment ||
		t == domain.ContributionTypeReviewComment ||
		t == domain.ContributionTypePRComment ||
		t == domain.ContributionTypeDiscussionComment
}
