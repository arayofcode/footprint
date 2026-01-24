package scoring

import (
	"fmt"
	"math"

	"github.com/arayofcode/footprint/internal/github"
)

const (
	PRScore                float64 = 10.0
	IssueScore             float64 = 5.0
	IssueCommentScore      float64 = 2.0
	ReviewScore            float64 = 3.0
	ReviewCommentScore     float64 = 2.0
	DiscussionScore        float64 = 2.0
	DiscussionCommentScore float64 = 2.0
)

var baseContributionScores = map[github.ContributionType]float64{
	github.ContributionTypePR:                PRScore,
	github.ContributionTypeIssue:             IssueScore,
	github.ContributionTypeIssueComment:      IssueCommentScore,
	github.ContributionTypeReview:            ReviewScore,
	github.ContributionTypeReviewComment:     ReviewCommentScore,
	github.ContributionTypeDiscussion:        DiscussionScore,
	github.ContributionTypeDiscussionComment: DiscussionCommentScore,
}

func baseScore(event github.ContributionEvent) float64 {
	if score, ok := baseContributionScores[event.Type]; ok {
		return score
	}
	fmt.Printf("Score not found for type: %s", event.Type)
	return 0
}

func popularityMultiplier(event github.ContributionEvent) float64 {
	// Formula: 1 + log10(1 + repo_stars + 2*repo_forks)
	val := 1.0 + float64(event.Stars) + 2.0*float64(event.Forks)
	return 1.0 + math.Log10(val)
}

func ImpactScore(event github.ContributionEvent) float64 {
	return baseScore(event) * popularityMultiplier(event)
}
