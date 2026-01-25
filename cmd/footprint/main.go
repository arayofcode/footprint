package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/arayofcode/footprint/internal/app"
)

func main() {
	var (
		username   string
		minStars   int
		outputDir  string
		clamp      float64
		timeout    time.Duration
		enableCard bool
	)
	flag.StringVar(&username, "username", "", "GitHub username (defaults to GITHUB_ACTOR)")
	flag.IntVar(&minStars, "min-stars", 5, "Minimum stars for owned projects")
	flag.StringVar(&outputDir, "output", "dist", "Output directory")
	flag.Float64Var(&clamp, "clamp", 0, "Clamp for popularity multiplier (0 = default)")
	flag.DurationVar(&timeout, "timeout", 60*time.Second, "Timeout for GitHub API operations")
	flag.BoolVar(&enableCard, "card", true, "Generate SVG card")
	flag.Parse()

	if err := app.RunCLI(context.Background(), app.CLIConfig{
		Username:   username,
		MinStars:   minStars,
		OutputDir:  outputDir,
		Clamp:      clamp,
		Timeout:    timeout,
		EnableCard: enableCard,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
