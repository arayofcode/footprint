package logic

import (
	"testing"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.ContributionEvent
		expected domain.SemanticEventType
	}{
		{
			name:     "PR Opened",
			input:    domain.ContributionEvent{Type: domain.ContributionTypePR},
			expected: domain.SemanticEventPrOpened,
		},
		{
			name:     "PR Review",
			input:    domain.ContributionEvent{Type: domain.ContributionTypeReview},
			expected: domain.SemanticEventPrReview,
		},
		{
			name:     "Issue Opened",
			input:    domain.ContributionEvent{Type: domain.ContributionTypeIssue},
			expected: domain.SemanticEventIssueOpened,
		},
		{
			name:     "Issue Comment (Genuine)",
			input:    domain.ContributionEvent{Type: domain.ContributionTypeIssueComment, URL: "https://github.com/a/b/issues/1#comment-1"},
			expected: domain.SemanticEventIssueComment,
		},
		{
			name:     "Issue Comment (Actually PR)",
			input:    domain.ContributionEvent{Type: domain.ContributionTypeIssueComment, URL: "https://github.com/a/b/pull/1#comment-1"},
			expected: domain.SemanticEventPrReviewComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.input)
			if got.Type != tt.expected {
				t.Errorf("Classify() type = %v, want %v", got.Type, tt.expected)
			}
		})
	}
}
