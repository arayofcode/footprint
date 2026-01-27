package scoring

import (
	"math"
	"testing"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestNewCalculator_DefaultClamp(t *testing.T) {
	calculator := NewCalculator(0)
	if calculator.Clamp != DefaultClamp {
		t.Fatalf("expected default clamp %v, got %v", DefaultClamp, calculator.Clamp)
	}
}

func TestScoreContribution_UsesBaseScoreAndPopularity(t *testing.T) {
	calculator := NewCalculator(0)
	event := domain.ContributionEvent{
		Type:  domain.ContributionTypePR,
		Stars: 100,
		Forks: 50,
	}

	scored := calculator.ScoreContribution(event)

	base := 10.0
	multiplier := 1 + math.Log10(1+float64(event.Stars)+2*float64(event.Forks))
	if multiplier > DefaultClamp {
		multiplier = DefaultClamp
	}
	expected := base * multiplier

	assertFloatApprox(t, expected, scored.Score, 1e-9)
}

func TestScoreContribution_MergedPRGetsBonus(t *testing.T) {
	calculator := NewCalculator(0)

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

	if scoredMerged.Score <= scoredUnmerged.Score {
		t.Fatalf("expected merged PR score to exceed unmerged score, got merged=%v unmerged=%v", scoredMerged.Score, scoredUnmerged.Score)
	}
}

func TestScoreOwnedProject_UsesPopularityMultiplier(t *testing.T) {
	calculator := NewCalculator(4.0)
	project := domain.OwnedProject{
		Stars: 100,
		Forks: 50,
	}

	scored := calculator.ScoreOwnedProject(project)

	expected := 1 + math.Log10(1+float64(project.Stars)+2*float64(project.Forks))
	if expected > calculator.Clamp {
		expected = calculator.Clamp
	}
	expected *= OwnershipScore

	assertFloatApprox(t, expected, scored.Score, 1e-9)
}

func assertFloatApprox(t *testing.T, expected, actual, tolerance float64) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Fatalf("expected %v, got %v (tolerance %v)", expected, actual, tolerance)
	}
}
