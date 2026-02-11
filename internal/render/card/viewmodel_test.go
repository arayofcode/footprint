package card

import (
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestBuildCardViewModel(t *testing.T) {
	user := domain.User{Username: "ray", AvatarURL: "https://example.com/avatar.png"}
	stats := domain.StatsView{
		PRsOpened:     5,
		PRReviews:     0,
		IssuesOpened:  10,
		IssueComments: 0,
		ProjectsOwned: 1,
		StarsEarned:   100,
	}
	generatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("Labels match expected display names", func(t *testing.T) {
		vm := buildViewModel(user, stats, generatedAt, nil, nil, true, false, false)
		expected := map[string]bool{
			"PRs Opened":     false,
			"Code Reviews":   false,
			"Issues Opened":  false,
			"Issue Comments": false,
			"Projects Owned": false,
			"Stars Earned":   false,
		}

		for _, s := range vm.Stats {
			if _, ok := expected[s.Label]; ok {
				expected[s.Label] = true
			} else {
				t.Errorf("Unexpected stat label: %s", s.Label)
			}
		}

		for label, found := range expected {
			if !found {
				t.Errorf("Expected stat label not found: %s", label)
			}
		}
	})

	t.Run("Zero-value stats excluded when showAllStats=false", func(t *testing.T) {
		vm := buildViewModel(user, stats, generatedAt, nil, nil, false, false, false)

		for _, s := range vm.Stats {
			if s.Raw == 0 {
				t.Errorf("Expected zero-value stat %s to be excluded", s.Label)
			}
		}

		if len(vm.Stats) != 3 {
			t.Errorf("Expected 3 stats, got %d", len(vm.Stats))
		}
	})

	t.Run("Sections omitted when empty and minimalSections=true", func(t *testing.T) {
		vm := buildViewModel(user, stats, generatedAt, nil, nil, false, true, true)
		if len(vm.Sections) != 0 {
			t.Errorf("Expected 0 sections, got %d", len(vm.Sections))
		}
	})

	t.Run("User Avatar Key matches", func(t *testing.T) {
		vm := buildViewModel(user, stats, generatedAt, nil, nil, true, false, false)
		expectedKey := domain.UserAvatarKey(user.Username)
		if vm.User.AvatarKey != expectedKey {
			t.Errorf("Expected avatar key %v, got %v", expectedKey, vm.User.AvatarKey)
		}
	})
}
