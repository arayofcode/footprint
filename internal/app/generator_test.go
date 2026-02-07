package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

type fakeFetcher struct {
	events []domain.ContributionEvent
	user   domain.User
	err    error
}

func (f fakeFetcher) FetchExternalContributions(ctx context.Context, username string) (domain.User, []domain.ContributionEvent, error) {
	if f.err != nil {
		return domain.User{}, nil, f.err
	}
	u := f.user
	if u.Username == "" {
		u.Username = username
	}
	return u, f.events, nil
}

type fakeProjects struct {
	projects []domain.OwnedProject
	err      error
}

func (f fakeProjects) FetchOwnedProjects(ctx context.Context, username string, minStars int) ([]domain.OwnedProject, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.projects, nil
}

type fakeScorer struct{}

func (fakeScorer) ScoreContribution(event domain.ContributionEvent) domain.ContributionEvent {
	event.Score = 42
	return event
}

func (fakeScorer) ScoreBatch(events []domain.ContributionEvent) []domain.ContributionEvent {
	scored := make([]domain.ContributionEvent, len(events))
	for i, e := range events {
		e.Score = 42
		scored[i] = e
	}
	return scored
}

func (fakeScorer) ScoreOwnedProject(project domain.OwnedProject) domain.OwnedProject {
	project.Score = 99
	return project
}

type fakeReportRenderer struct {
	user        domain.User
	stats       domain.UserStats
	generatedAt time.Time
	events      []domain.ContributionEvent
	projects    []domain.OwnedProject
	err         error
}

func (f *fakeReportRenderer) RenderReport(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	f.user = user
	f.stats = stats
	f.generatedAt = generatedAt
	f.events = events
	f.projects = projects
	if f.err != nil {
		return nil, f.err
	}
	return []byte("report"), nil
}

type fakeSummaryRenderer struct {
	user        domain.User
	stats       domain.UserStats
	generatedAt time.Time
	events      []domain.ContributionEvent
	projects    []domain.OwnedProject
	err         error
}

func (f *fakeSummaryRenderer) RenderSummary(ctx context.Context, user domain.User, stats domain.UserStats, generatedAt time.Time, events []domain.ContributionEvent, projects []domain.OwnedProject) ([]byte, error) {
	f.user = user
	f.stats = stats
	f.generatedAt = generatedAt
	f.events = events
	f.projects = projects
	if f.err != nil {
		return nil, f.err
	}
	return []byte("summary"), nil
}

type fakeCardRenderer struct {
	called bool
	err    error
}

func (f *fakeCardRenderer) RenderCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProject, assets map[string]string) ([]byte, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return []byte("card"), nil
}

func (f *fakeCardRenderer) RenderMinimalCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProject, assets map[string]string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("minimal-card"), nil
}

func (f *fakeCardRenderer) RenderExtendedCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProject, assets map[string]string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("extended-card"), nil
}

func (f *fakeCardRenderer) RenderExtendedMinimalCard(ctx context.Context, user domain.User, stats domain.StatsView, generatedAt time.Time, contributions []domain.RepoContribution, projects []domain.OwnedProject, assets map[string]string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("extended-minimal-card"), nil
}

type fakeWriter struct {
	writes map[string][]byte
	err    error
}

func (f *fakeWriter) Write(ctx context.Context, filename string, data []byte) error {
	if f.err != nil {
		return f.err
	}
	if f.writes == nil {
		f.writes = make(map[string][]byte)
	}
	f.writes[filename] = data
	return nil
}

func TestGeneratorRun_Success_NoCard(t *testing.T) {
	fetcher := fakeFetcher{
		events: []domain.ContributionEvent{
			{ID: "1", Type: domain.ContributionTypePR, Repo: "a/b", Stars: 10, Forks: 2},
		},
	}
	projects := fakeProjects{
		projects: []domain.OwnedProject{
			{Repo: "me/owned", Stars: 50, Forks: 5},
		},
	}
	scorer := fakeScorer{}
	reportRenderer := &fakeReportRenderer{}
	summaryRenderer := &fakeSummaryRenderer{}
	writer := &fakeWriter{}

	gen := &Generator{
		Fetcher:         fetcher,
		Projects:        projects,
		Scorer:          scorer,
		ReportRenderer:  reportRenderer,
		SummaryRenderer: summaryRenderer,
		Writer:          writer,
		MinStars:        5,
	}

	if err := gen.Run(context.Background(), "ray"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := writer.writes["report.json"]; !ok {
		t.Fatalf("expected report.json to be written")
	}
	if _, ok := writer.writes["summary.md"]; !ok {
		t.Fatalf("expected summary.md to be written")
	}

	if reportRenderer.user.Username != "ray" {
		t.Fatalf("expected report renderer username ray, got %s", reportRenderer.user.Username)
	}
	if reportRenderer.generatedAt.IsZero() {
		t.Fatalf("expected report renderer generatedAt to be set")
	}
	if len(reportRenderer.events) != 1 || reportRenderer.events[0].Score != 42 {
		t.Fatalf("expected scored events to be passed to report renderer")
	}
	if len(reportRenderer.projects) != 1 || reportRenderer.projects[0].Score != 99 {
		t.Fatalf("expected scored projects to be passed to report renderer")
	}

	if summaryRenderer.user.Username != "ray" {
		t.Fatalf("expected summary renderer username ray, got %s", summaryRenderer.user.Username)
	}
	if summaryRenderer.generatedAt.IsZero() {
		t.Fatalf("expected summary renderer generatedAt to be set")
	}
}

func TestGeneratorRun_WithCard(t *testing.T) {
	gen := &Generator{
		Fetcher:         fakeFetcher{},
		Projects:        fakeProjects{},
		Scorer:          fakeScorer{},
		ReportRenderer:  &fakeReportRenderer{},
		SummaryRenderer: &fakeSummaryRenderer{},
		CardRenderer:    &fakeCardRenderer{},
		Writer:          &fakeWriter{},
	}

	if err := gen.Run(context.Background(), "ray"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := gen.Writer.(*fakeWriter).writes["card.svg"]; !ok {
		t.Fatalf("expected card.svg to be written")
	}
}

func TestGeneratorRun_DependencyMissing(t *testing.T) {
	gen := &Generator{
		Fetcher:         fakeFetcher{},
		Projects:        fakeProjects{},
		ReportRenderer:  &fakeReportRenderer{},
		SummaryRenderer: &fakeSummaryRenderer{},
		Writer:          &fakeWriter{},
	}

	if err := gen.Run(context.Background(), "ray"); err == nil {
		t.Fatalf("expected error due to missing dependencies")
	}
}

func TestGeneratorRun_PropagatesFetchError(t *testing.T) {
	expectedErr := errors.New("fetch failed")
	gen := &Generator{
		Fetcher:         fakeFetcher{err: expectedErr},
		Projects:        fakeProjects{},
		Scorer:          fakeScorer{},
		ReportRenderer:  &fakeReportRenderer{},
		SummaryRenderer: &fakeSummaryRenderer{},
		Writer:          &fakeWriter{},
	}

	if err := gen.Run(context.Background(), "ray"); err == nil {
		t.Fatalf("expected fetch error to be returned")
	}
}

func TestGeneratorRun_PropagatesRenderError(t *testing.T) {
	expectedErr := errors.New("render failed")
	gen := &Generator{
		Fetcher:         fakeFetcher{},
		Projects:        fakeProjects{},
		Scorer:          fakeScorer{},
		ReportRenderer:  &fakeReportRenderer{err: expectedErr},
		SummaryRenderer: &fakeSummaryRenderer{},
		Writer:          &fakeWriter{},
	}

	if err := gen.Run(context.Background(), "ray"); err == nil {
		t.Fatalf("expected render error to be returned")
	}
}
