package card

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

func TestRenderCard_GeneratesSVGWithStats(t *testing.T) {
	renderer := Renderer{}
	events := []domain.RepoContribution{
		{Repo: "a/b", Score: 7},
	}
	projects := []domain.OwnedProject{
		{Repo: "me/owned", Stars: 10, Forks: 2},
	}
	user := domain.User{Username: "ray", AvatarURL: "https://example.com/avatar.png"}
	stats := domain.StatsView{
		PRsOpened:     5,
		PRReviews:     3,
		IssuesOpened:  2,
		IssueComments: 10,
	}

	out, err := renderer.RenderCard(context.Background(), user, stats, time.Now(), events, projects, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Check for SVG header
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG output")
	}

	// Check for landscape dimensions (800px wide)
	if !strings.Contains(svg, `width="800"`) {
		t.Error("expected landscape width of 800")
	}

	// Check for new stats labels
	expectedLabels := []string{
		"PRS OPENED",
		"PRS REVIEWED",
		"ISSUES OPENED",
		"COMMENTS MADE",
		"PROJECTS OWNED",
		"STARS EARNED",
	}
	for _, label := range expectedLabels {
		if !strings.Contains(svg, label) {
			t.Errorf("expected SVG to contain label %q", label)
		}
	}

	// Check for values
	if !strings.Contains(svg, ">5<") { // Total PRs
		t.Error("expected SVG to contain PR count 5")
	}
	if !strings.Contains(svg, ">10<") { // Stars (from projects)
		t.Error("expected SVG to contain Star count 10")
	}

	// Standard card should NOT contain sections
	if strings.Contains(svg, "TOP PROJECTS CREATED") {
		t.Error("standard card should not contain 'TOP PROJECTS CREATED' section")
	}
	if strings.Contains(svg, "KEY CONTRIBUTIONS") {
		t.Error("standard card should not contain 'KEY CONTRIBUTIONS' section")
	}
}

func TestRenderMinimalCard_HidesZeroStats(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{
		PRsOpened:     5,
		PRReviews:     0, // Zero - should be hidden
		IssuesOpened:  2,
		IssueComments: 0, // Zero - should be hidden
	}

	out, err := renderer.RenderMinimalCard(context.Background(), user, stats, time.Now(), nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Should contain non-zero labels
	if !strings.Contains(svg, "PRS OPENED") {
		t.Error("expected SVG to contain 'PRS OPENED'")
	}
	if !strings.Contains(svg, "ISSUES OPENED") {
		t.Error("expected SVG to contain 'ISSUES OPENED'")
	}

	// Should NOT contain zero-value labels
	if strings.Contains(svg, "PRS REVIEWED") {
		t.Error("expected SVG to hide zero-value 'PRS REVIEWED'")
	}
	if strings.Contains(svg, "COMMENTS MADE") {
		t.Error("expected SVG to hide zero-value 'COMMENTS MADE'")
	}
}

func TestRenderExtendedCard_IncludesSections(t *testing.T) {
	renderer := Renderer{}
	events := []domain.RepoContribution{
		{Repo: "external/repo", Score: 5, AvatarURL: "https://example.com/avatar.png"},
	}
	projects := []domain.OwnedProject{
		{Repo: "myrepo", Stars: 10, Forks: 2, URL: "https://github.com/ray/myrepo"},
	}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{PRsOpened: 5}

	out, err := renderer.RenderExtendedCard(context.Background(), user, stats, time.Now(), events, projects, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Should contain sections
	if !strings.Contains(svg, "TOP PROJECTS CREATED") {
		t.Error("expected SVG to contain 'TOP PROJECTS CREATED' section")
	}
	if !strings.Contains(svg, "KEY CONTRIBUTIONS") {
		t.Error("expected SVG to contain 'KEY CONTRIBUTIONS' section")
	}
}

func TestRenderExtendedMinimalCard_HidesEmptySections(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{PRsOpened: 5}

	// No events and no projects - sections should be hidden
	out, err := renderer.RenderExtendedMinimalCard(context.Background(), user, stats, time.Now(), nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Should NOT contain sections when empty
	if strings.Contains(svg, "TOP PROJECTS CREATED") {
		t.Error("expected SVG to hide 'TOP PROJECTS CREATED' when no projects")
	}
	if strings.Contains(svg, "KEY CONTRIBUTIONS") {
		t.Error("expected SVG to hide 'KEY CONTRIBUTIONS' when no external contributions")
	}
}

func TestRenderExtendedMinimalCard_ShiftsExternalToLeft(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{PRsOpened: 5}
	events := []domain.RepoContribution{
		{Repo: "ext/repo", Score: 5},
	}

	// No projects, but has events. Key Contributions should move to x=40.
	out, err := renderer.RenderExtendedMinimalCard(context.Background(), user, stats, time.Now(), events, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	// Should contain Key Contributions at x=40
	if !strings.Contains(svg, `<g transform="translate(40,`) {
		t.Error("expected section to be at x=40 when left side is empty")
	}
	if !strings.Contains(svg, "KEY CONTRIBUTIONS") {
		t.Error("expected SVG to contain 'KEY CONTRIBUTIONS'")
	}
}

func TestRenderCard_SVGContainsCoreSections(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{PRsOpened: 5}

	out, err := renderer.RenderCard(context.Background(), user, stats, time.Now(), nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	svg := string(out)

	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG root")
	}
	if !strings.Contains(svg, "PRS OPENED") {
		t.Error("expected stat label PRS OPENED")
	}
}

func TestRenderCard_IsDeterministic(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray", AvatarURL: "https://example.com/avatar.png"}
	stats := domain.StatsView{PRsOpened: 5, IssuesOpened: 2}
	events := []domain.RepoContribution{
		{Repo: "a/b", Score: 10},
		{Repo: "c/d", Score: 5},
	}
	// Mock assets ensuring stable input
	assets := map[domain.AssetKey]string{
		domain.UserAvatarKey("ray"): "data:image/png;base64,AAAA",
	}
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	out1, err := renderer.RenderCard(context.Background(), user, stats, now, events, nil, assets)
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	out2, err := renderer.RenderCard(context.Background(), user, stats, now, events, nil, assets)
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	if string(out1) != string(out2) {
		t.Error("RenderCard output is not deterministic")
	}
}

func TestRegression_Labels(t *testing.T) {
	renderer := Renderer{}
	user := domain.User{Username: "ray"}
	stats := domain.StatsView{PRReviews: 5}

	out, err := renderer.RenderCard(context.Background(), user, stats, time.Now(), nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	svg := string(out)

	if strings.Contains(svg, "PR FEEDBACK") {
		t.Error("SVG should NOT contain legacy label 'PR FEEDBACK'")
	}
	if !strings.Contains(svg, "PRS REVIEWED") {
		t.Error("SVG SHOULD contain label 'PRS REVIEWED'")
	}
}
func TestRenderSection_DoesNotInferFromBadges(t *testing.T) {
	// If RowKind is External, even with a Star icon it should use external layout
	layout := LayoutVM{RowHeight: 55, Width: 800}
	assets := map[domain.AssetKey]string{}
	resolve := func(k domain.AssetKey) string { return assets[k] }

	sec := SectionVM{
		Title: "TEST",
		Rows: []SectionRowVM{
			{
				Kind: RowExternalContribution,
				Badges: []BadgeVM{
					{Icon: iconStar, Count: "10"},
				},
			},
		},
	}

	svg := renderSection(sec, layout, resolve)

	// External layout uses currentX := -50 and translates by cardWidth-5 (340-5=335 or 420-5=415)
	// Owned layout uses translate cardWidth-80 (340-80=260 or 420-80=340)

	// Since cards are 340 wide in horizontal (default),
	// Owned would be at 260.
	// External would be at 335.

	if strings.Contains(svg, `translate(260, 10.5)`) {
		t.Error("expected external layout for RowExternalContribution, but got owned layout (translate 260)")
	}
	if !strings.Contains(svg, `translate(335, 10.5)`) {
		t.Error("expected external layout (translate 335) for RowExternalContribution")
	}
}

func TestRenderSection_ShowsEmptyMessage(t *testing.T) {
	layout := LayoutVM{Width: 800}
	sec := SectionVM{
		Title:        "EMPTY",
		EmptyMessage: "Nothing here",
		Rows:         nil,
	}

	svg := renderSection(sec, layout, nil)
	if !strings.Contains(svg, "Nothing here") {
		t.Error("expected empty message to be rendered")
	}
}
