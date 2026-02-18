# Footprint

Footprint is a GitHub Action and tool that discovers a user’s public open-source contributions across GitHub (beyond just PRs) and generates a portfolio-ready “footprint” card + report. It's designed to help developers showcase their impact, especially across repositories they don't own.

## Features

- **Deep Discovery**: Finds contributions across GitHub, focusing on impact beyond just code ownership.
    - **Contributions**: PRs opened, Code Reviews (PR comments/ reviews), Issues opened, and Issue comments.
    - **Highlighting**: Top external contributions (ranked by impact) + owned repo highlights (ranked by stars).
- **Impact & Quality**:
    - **Merged Bonus**: Merged Pull Requests receive a **1.5x bonus** to prioritize landing code.
    - **Repo-Level Popularity Multiplier**: Scores are scaled using `1 + log10(1 + stars + 2*forks)` to reward impact in widely-adopted projects. This multiplier is applied **once per repository** and capped at 4.0, preventing star-heavy repos from dominating via sheer volume of low-impact events.
    - **Diminishing Returns**: Comment/Issue scores decay per repository using a `1.0 / (1.0 + 0.5 * count)` formula. This ensures that while consistent engagement is valued, repetitive low-effort activity has reduced impact compared to diverse, high-value contributions.
    - **Multi-Metric Separation**: Strictly separates PRs Opened, PR Reviews, PR Comments, Issues Opened, and Issue Comments to prevent conflation.
    - **Minimal Card Expansion**: In minimal layouts, if one section (Owned or External) is empty, the other expands to show up to 6 items. To utilize space effectively, these items "shift to the right" into a **2x3 grid**, occupying the remaining horizontal space.
- **Card Variants**:
    - **Standard**: All core statistics.
      ![Standard Card](/docs/images/card-standard.svg)
    - **Minimal**: Non-zero statistics only.
      ![Minimal Card](/docs/images/card-minimal.svg)
    - **Extended**: Rich dashboard including top projects and contribution highlights.
      ![Extended Card](/docs/images/card-extended.svg)
    - **Extended Minimal**: Minified dashboard view.
      ![Extended Minimal Card](/docs/images/card-extended-minimal.svg)

- **Artifact Generation**:
    - `summary.md`: Human-readable impact summary.
    - `report.json`: Machine-readable scoring data.
    - `card.svg` (+ variants): Interactive SVG visualizations.

## SVG Embedding

To showcase your footprint on your GitHub profile, embed the generated cards using Markdown or HTML:

### Markdown
```markdown
![Footprint Card](https://raw.githubusercontent.com/<username>/<repo>/output/card.svg)
```

### HTML (for links)
```html
<a href="https://github.com/<username>/<repo>">
  <img src="https://raw.githubusercontent.com/<username>/<repo>/output/card.svg" alt="Footprint" width="800" />
</a>
```

> [!TIP]
> Use the `output` branch (or similar) to host these artifacts for clean embedding.

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
```

Environment variables:

```bash
- `GITHUB_TOKEN` (required)
- `GITHUB_ACTOR` (used when `-username` is not provided)

## GitHub Action Usage

Add this to your `username/username` repository (or any repo) as a workflow:

```yaml
# .github/workflows/footprint.yml
name: Generate Footprint
on:
  schedule:
    - cron: "0 0 * * 0" # Weekly
  workflow_dispatch:

permissions:
  contents: write # for pushing artifacts to output branch

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Generate Footprint
        uses: arayofcode/footprint@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          min_stars: 5
          card: true
```

## Architecture Principles

- **Semantic ViewModel**: The renderer receives data structures (`RowKind`, `Badges`) rather than raw URLs or domain inferences. ViewModels contain no geometry (X/Y) or layout-specific logic.
- **Centralized Layout**: All geometric math (X, Y, height, gaps) is pre-calculated in a pure `DecideLayout` function. `LayoutVM` exclusively owns all coordinates.
- **Zero Heuristics**: SVG generation logic is purely compositional and branching is based on explicit ViewModel flags.
- **Purely Deterministic**: Given the same ViewModel and assets, the SVG output is byte-perfect identical every time.

## Roadmap & Areas for Improvement

- **New Data Sources**: Wiki edits, PR comments, Gist activity, and GitHub Discussions.
- **Detailed Analytics**: Refine scoring using metadata like `merged by`, `pushed to main`, or `tagged as help wanted`.
- **Reaction Indexing**: Incorporate emoji reactions for qualitative impact assessment.
