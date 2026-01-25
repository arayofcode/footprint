package domain

import (
	"math"
	"testing"
)

func TestPopularityMultiplierClamp(t *testing.T) {
	event := ContributionEvent{
		Stars: 10000,
		Forks: 5000,
	}

	clamp := 4.0
	multiplier := event.PopularityMultiplier(clamp)

	if multiplier > clamp {
		t.Fatalf("expected multiplier to be clamped to %v, got %v", clamp, multiplier)
	}

	uncapped := 1 + math.Log10(1+float64(event.Stars)+2*float64(event.Forks))
	if uncapped <= clamp {
		t.Fatalf("expected uncapped multiplier to exceed clamp, got %v", uncapped)
	}
}
