package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/arayofcode/footprint/internal/github"
	"github.com/arayofcode/footprint/internal/render"
	"github.com/arayofcode/footprint/internal/scoring"
)

const outputDir = "dist"

func main() {
	ctx := context.Background()

	username := os.Getenv("GITHUB_ACTOR")
	if username == "" {
		username = "Bhupesh-V"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required for GraphQL API")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := github.NewClient(githubv4.NewClient(httpClient))

	log.Printf("Fetching PRs for user: %s\n", username)

	// PRs created in non self-owned repos
	externalPRs, err := client.FetchExternalPRs(ctx, username)
	if err != nil {
		log.Fatalf("Error fetching external PRs: %v", err)
	}
	log.Printf("Found %d external PRs\n", len(externalPRs))

	// PRs in self-owned repos with >5 stars
	ownRepoPRs, err := client.FetchOwnRepoPRs(ctx, username, 5)
	if err != nil {
		log.Fatalf("Error fetching own repo PRs: %v", err)
	}
	log.Printf("Found %d PRs from own repos (>5 stars)\n", len(ownRepoPRs))

	allEvents := append(externalPRs, ownRepoPRs...)
	log.Printf("Total PRs: %d\n", len(allEvents))

	// Calculate scores
	for _, event := range allEvents {
		event.Score = scoring.ImpactScore(*event)
	}

	report := render.GenerateReport(username, allEvents)

	log.Printf("Writing artifacts to %s/\n", outputDir)

	if err := render.WriteReportJSON(report, outputDir); err != nil {
		log.Fatalf("Error writing report.json: %v", err)
	}
	log.Println("Generated dist/report.json")

	if err := render.WriteSummaryMarkdown(report, outputDir); err != nil {
		log.Fatalf("Error writing summary.md: %v", err)
	}
	log.Println("Generated dist/summary.md")

	output, err := json.MarshalIndent(allEvents, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Println(string(output))
}
