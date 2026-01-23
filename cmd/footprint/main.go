package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	ghlib "github.com/google/go-github/v81/github"

	"github.com/arayofcode/footprint/internal/github"
	"github.com/arayofcode/footprint/internal/render"
)

const outputDir = "dist"

func main() {
	ctx := context.Background()

	username := os.Getenv("GITHUB_ACTOR")
	if username == "" {
		username = "Bhupesh-V"
	}
	var ghClient *ghlib.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ghClient = ghlib.NewClient(nil).WithAuthToken(token)
		log.Println("Using authenticated GitHub client")
	} else {
		ghClient = ghlib.NewClient(nil)
		log.Println("Using unauthenticated GitHub client (lower rate limits)")
	}

	client := github.NewClient(ghClient)

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
