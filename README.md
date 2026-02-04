# Footprint

Footprint is a GitHub Action and tool that discovers a user’s public open-source contributions across GitHub (beyond just PRs) and generates a portfolio-ready “footprint” card + report. It's designed to help developers showcase their impact, especially across repositories they don't own.

## Features

- **Deep Discovery**: Finds contributions across GitHub, focusing on impact beyond just code ownership.
    - **Contributions**: PRs opened, PR feedback (comments/ reviews), Issues opened, and Issue comments.
    - **Highlighting**: Top external contributions (ranked by impact) + owned repo highlights (ranked by stars).
- **Impact & Quality**:
    - **Merged Bonus**: Merged Pull Requests receive a **1.5x bonus** to prioritize landing code.
    - **Popularity Multiplier**: Scores are scaled using `1 + log10(1 + stars + 2*forks)` to reward impact in widely-adopted projects.
    - **Repo-Level Aggregation**: Ranks repositories by your **Total Impact Score** across all contributions.
- **Artifact Generation**:
    - `dist/summary.md`: Human-readable portfolio summary.
    - `dist/report.json`: JSON version of your footprint.
    - `dist/card.svg`: Dynamic, interactive SVG card clickable stats, taking you to exact contributions.
    - `card-extended.svg`: Rich dashboard view including top projects and external contribution highlights.

## Usage

Run locally:

```bash
go run ./cmd/footprint
```

Optional flags:

```bash
-username <github_username>
-min-stars <n>
-output <dir>
-clamp <multiplier_cap>
```

Environment variables:

- `GITHUB_TOKEN` (required)
- `GITHUB_ACTOR` (used when `-username` is not provided)

## GitHub Action Usage

Add this to your `username/username` repository (or any repo) as a workflow:

```yaml
# .github/workflows/footprint.yml
name: Generate Footprint
on:
  schedule:
    - cron: '0 0 * * 0' # Weekly
  workflow_dispatch:

jobs:
  footprint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Generate Footprint
        id: footprint
        uses: arayofcode/footprint@main
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      # The footprint summary will now appear in your Action Run UI!
      # You can also use the outputs:
      - name: Print Score
        run: echo "Calculated Score: ${{ steps.footprint.outputs.total_score }}"
```

## Roadmap & Areas for Improvement

- **New Data Sources**: Wiki edits, PR comments, Gist activity, and GitHub Discussions.
- **Detailed Analytics**: Refine scoring using metadata like `merged by`, `pushed to main`, or `tagged as help wanted`.
- **Reaction Indexing**: Incorporate emoji reactions for qualitative impact assessment.
