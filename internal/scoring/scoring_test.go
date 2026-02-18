package scoring

import (
	"math"
	"testing"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestScoreContribution_UsesBaseScoreAndPopularity(t *testing.T) {
	calculator := NewCalculator()
	event := domain.ContributionEvent{
		Type:  domain.ContributionTypePR,
		Stars: 100,
		Forks: 50,
	}

	scored := calculator.ScoreContribution(event)

	expectedBase := 10.0
	expectedPopularity := 1 + math.Log10(1+float64(event.Stars)+2*float64(event.Forks))

	assertFloatApprox(t, expectedBase, scored.BaseScore, 1e-9)
	assertFloatApprox(t, expectedPopularity, scored.PopularityRaw, 1e-9)
}

func TestScoreContribution_MergedPRGetsBonus(t *testing.T) {
	calculator := NewCalculator()

	merged := domain.ContributionEvent{
		Type:   domain.ContributionTypePR,
		Merged: true,
		Stars:  100,
		Forks:  50,
	}

	unmerged := domain.ContributionEvent{
		Type:   domain.ContributionTypePR,
		Merged: false,
		Stars:  100,
		Forks:  50,
	}

	scoredMerged := calculator.ScoreContribution(merged)
	scoredUnmerged := calculator.ScoreContribution(unmerged)

	// Merged bonus (1.5x) should be in BaseScore
	expectedBonusRatio := 1.5
	if math.Abs(scoredMerged.BaseScore/scoredUnmerged.BaseScore-expectedBonusRatio) > 1e-9 {
		t.Fatalf("expected merged base score to be 1.5x higher, got merged=%v unmerged=%v", scoredMerged.BaseScore, scoredUnmerged.BaseScore)
	}

	// Popularity should be identical
	if scoredMerged.PopularityRaw != scoredUnmerged.PopularityRaw {
		t.Fatalf("expected same popularity, got merged=%v unmerged=%v", scoredMerged.PopularityRaw, scoredUnmerged.PopularityRaw)
	}
}

func TestEnrichOwnedProject_PopulatesBaseAndPopularity(t *testing.T) {
	calculator := NewCalculator()
	project := domain.OwnedProject{
		Stars: 100,
		Forks: 50,
	}

	enriched := calculator.EnrichOwnedProject(project)

	expectedPopularity := 1 + math.Log10(1+float64(project.Stars)+2*float64(project.Forks))

	assertFloatApprox(t, OwnershipScore, enriched.BaseScore, 1e-9)
	assertFloatApprox(t, expectedPopularity, enriched.PopularityRaw, 1e-9)
}

func TestScoreBatch_AppliesDecay(t *testing.T) {
	calculator := NewCalculator()
	repo := "org/repo"

	// Create multiple comments in the same repo
	events := []domain.ContributionEvent{
		{Type: domain.ContributionTypeIssueComment, Repo: repo, Stars: 0, Forks: 0}, // 1st: 1.0x
		{Type: domain.ContributionTypeIssueComment, Repo: repo, Stars: 0, Forks: 0}, // 2nd: 0.66x
		{Type: domain.ContributionTypeIssueComment, Repo: repo, Stars: 0, Forks: 0}, // 3rd: 0.5x
	}

	scored := calculator.ScoreBatch(events)

	// Base for comment is 2.0
	// 1st: 2 * 1 * (1.0 / (1 + 0.5*0)) = 2.0
	// 2nd: 2 * 1 * (1.0 / (1 + 0.5*1)) = 1.333...
	// 3rd: 2 * 1 * (1.0 / (1 + 0.5*2)) = 1.0

	assertFloatApprox(t, 2.0, scored[0].BaseScore, 1e-9)
	assertFloatApprox(t, 1.333333333, scored[1].BaseScore, 1e-9)
	assertFloatApprox(t, 1.0, scored[2].BaseScore, 1e-9)
}

func assertFloatApprox(t *testing.T, expected, actual, tolerance float64) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Fatalf("expected %v, got %v (tolerance %v)", expected, actual, tolerance)
	}
}
