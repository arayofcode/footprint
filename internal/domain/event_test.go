package domain

import (
	"math"
	"testing"
)

func TestPopularityMultiplier(t *testing.T) {
	event := ContributionEvent{
		Stars: 100,
		Forks: 50,
	}

	expected := 1 + math.Log10(1+float64(event.Stars)+2*float64(event.Forks))
	actual := event.PopularityMultiplier()

	if math.Abs(expected-actual) > 1e-9 {
		t.Fatalf("expected multiplier %v, got %v", expected, actual)
	}
}

func TestStableID(t *testing.T) {
	e := ContributionEvent{ID: "fixed"}
	if e.StableID() != "fixed" {
		t.Errorf("expected fixed ID, got %s", e.StableID())
	}

	e2 := ContributionEvent{Repo: "a/b", URL: "https://github.com/a/b/pull/1"}
	expected := "a/b#https://github.com/a/b/pull/1"
	if e2.StableID() != expected {
		t.Errorf("expected generated ID %s, got %s", expected, e2.StableID())
	}
}
