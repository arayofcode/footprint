package scoring

import (
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
	// Sort by CreatedAt to ensure deterministic decay order
	// We operate on a copy or inplace? The interface implies returning a new slice or modified slice.
	// Let's modify in place but return the slice.
	// Actually better to sort a copy to avoid side effects if caller relies on order?
	// But usually we want chronological order anyway.
	// For now, let's assume events are mixed.
	// We need to group by repo and type for decay.

	// Map to track counts: repo -> type -> count
	counts := make(map[string]map[domain.ContributionType]int)

	scored := make([]domain.ContributionEvent, len(events))
	copy(scored, events)

	// Sort by date (oldest first) so early contributions get full value
	// We need a stable sort?
	// Let's rely on the fact that we process them in temporal order.
	// Check if we need to import sort.
	// For now, let's assuming strict temporal processing if we sort first.
	// BUT, we need to import "sort" first.
	// I'll add the method first, then add imports if needed.
	// actually I should check imports in scoring.go.

	for i := range scored {
		e := &scored[i]

		repo := e.Repo
		if _, ok := counts[repo]; !ok {
			counts[repo] = make(map[domain.ContributionType]int)
		}

		// Increment count for this type in this repo
		// We only decay comments
		count := counts[repo][e.Type]
		counts[repo][e.Type]++

		// Base score from single event scoring
		// We can reuse ScoreContribution but we need to inject the decay multiplier.
		// ScoreContribution calculates base * popularity * merged.
		// We want (base * decay) * popularity * merged.
		// So effectively we can calculate standard score, then apply decay multiplier.

		c.ScoreContribution(*e) // This mutates e.Score
		// Wait, ScoreContribution signature is (event) event, so it returns a copy?
		*e = c.ScoreContribution(*e)

		if isDecayable(e.Type) {
			decay := 1.0 / (1.0 + 0.5*float64(count)) // 1, 0.66, 0.5, 0.4...
			e.BaseScore *= decay
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
