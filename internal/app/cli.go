package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/arayofcode/footprint/internal/github"
	"github.com/arayofcode/footprint/internal/output"
	"github.com/arayofcode/footprint/internal/render/card"
	"github.com/arayofcode/footprint/internal/render/report"
	"github.com/arayofcode/footprint/internal/render/summary"
	"github.com/arayofcode/footprint/internal/scoring"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type CLIConfig struct {
	Username string

	MinStars  int
	OutputDir string
	Timeout   time.Duration

	EnableCard bool
}

func RunCLI(ctx context.Context, cfg CLIConfig) error {
	if ctx == nil {
		ctx = context.Background()
	}

	username := cfg.Username
	if username == "" {
		username = os.Getenv("GITHUB_ACTOR")
	}
	if username == "" {
		return fmt.Errorf("username is required (set CLIConfig.Username or GITHUB_ACTOR)")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is required for GitHub API access")
	}

	minStars := max(cfg.MinStars, 0)

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = "dist"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, src)
	ghClient := githubv4.NewClient(httpClient)

	client := github.NewClient(ghClient)
	writer := output.NewFileSystemWriter(outputDir)

	gen := &Generator{
		Fetcher:         client,
		Projects:        client,
		Scorer:          scoring.NewCalculator(),
		ReportRenderer:  report.Renderer{},
		SummaryRenderer: summary.Renderer{},
		Writer:          writer,
		Actions:         github.NewActions(),
		MinStars:        minStars,
	}

	if cfg.EnableCard {
		gen.CardRenderer = card.Renderer{}
	}

	if err := gen.Run(ctx, username); err != nil {
		return fmt.Errorf("footprint failed: %w", err)
	}

	return nil
}
