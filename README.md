# Footprint - *OSS contributions you forgot about*

Footprint is a GitHub Action and tool that discovers a user’s public open-source contributions across GitHub (beyond just PRs) and generates a portfolio-ready “footprint” card + report. It's designed to help developers showcase their impact, especially across repositories they don't own.

## Features

-   **Deep Discovery**: Finds contributions across GitHub, focusing on impact beyond just code ownership.
    -   *v0 Focus*: **PRs authored** in external repos (ranked by impact) + owned repo highlights (ranked by stars).
    -   *Future*: Issues, comments, reviews, discussions, and wiki edits.
-   **Impact & Quality**:
    -   Base-score + repo popularity multiplier (stars + forks).
    -   Repo-level aggregation: Ranks external projects by your **Total Impact Score**.
    -   Noise Filtering: Excludes owned repos below star threshold; omits empty card sections.
-   **Artifact Generation**:
    -   `dist/summary.md`: Human-readable portfolio summary.
    -   `dist/report.json`: JSON version of your footprint.
    -   `dist/card.svg`: Dynamic, interactive SVG card (v0 model).

## Usage

Run locally:

```/dev/null/usage.sh#L1-2
go run ./cmd/footprint
go run .
```

Optional flags:

```/dev/null/usage.sh#L1-5
-username <github_username>
-min-stars <n>
-output <dir>
-clamp <multiplier_cap>
```

Environment variables:

-   `GITHUB_TOKEN` (required)
-   `GITHUB_ACTOR` (used when `-username` is not provided)

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

### GitHub Environment Features
- **Job Summary**: Automatically writes the human-readable summary to the Actions Run Summary page.
- **Outputs**: Sets `total_contributions`, `owned_projects_count`, and `total_score` for use in other steps.

## Project Structure (high level)

-   `internal/domain`: Core types and interfaces.
-   `internal/github`: GitHub GraphQL fetchers (external contributions + owned repos).
-   `internal/scoring`: Base scores + popularity multiplier.
-   `internal/render/*`: Report, summary, and SVG renderers.
-   `internal/output`: Output writers (filesystem).
-   `internal/app`: Generator orchestration.


## Roadmap & Areas for Improvement

### Features
-   [x] **SVG Card**: Generate a dynamic `dist/card.svg` with impact-based ranking and clickable rows.
-   [ ] **More Contribution Types**: Add support for Issues, Issue Comments, Reviews, and Discussions.
-   [x] **Impact Scoring**: Base weights + popularity multiplier with repository-level aggregation.
-   [x] **Merged Detection**: accurately detect if a PR was merged.

### Code & Tech Debt
-   **Error Handling**: Improve granularity of error reporting in the GitHub client.
-   **Testing**: Add unit tests for the parser and renderer logic.
-   **Pagination**: optimize the search to handle extremely large contribution histories efficiently.
-   **Configuration**: Add support for a config file (e.g., `.footprint.yml`) to customize aggregation rules and ignore lists.
