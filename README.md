# Footprint - *OSS contributions you forgot about*

Footprint is a GitHub Action and tool that discovers a user’s public open-source contributions across GitHub (beyond just PRs) and generates a portfolio-ready “footprint” card + report. It's designed to help developers showcase their impact, especially across repositories they don't own.

## Features

-   **Deep Discovery**: Finds contributions across GitHub, focusing on impact beyond just code ownership.
    -   *v0 Focus*: Public Pull Requests in external repos + own popular repos.
    -   *Future*: Issues, helpful comments, reviews, discussions, and wiki edits.
-   **Impact & Quality**:
    -   Filters out noise (e.g., only includes own repos if >5 stars).
    -   *Future*: Heuristic scoring for "helpful" actions (e.g., comments marked as answers, highly reacted-to discussions).
-   **Artifact Generation**:
    -   `dist/summary.md`: Human-readable portfolio summary.
    -   `dist/report.json`: JSON version of your footprint.

## Roadmap & Areas for Improvement

### Features
-   [ ] **SVG Card**: Generate a `dist/card.svg` for embedding in READMEs (like `snk`).
-   [ ] **More Contribution Types**: Add support for Issues, Issue Comments, Reviews, and Discussions.
-   [ ] **Impact Scoring**: Implement a heuristic scoring system.
-   [ ] **Merged Detection**: accurately detect if a PR was merged (requires extra API checks).

### Code & Tech Debt
-   **Error Handling**: Improve granularity of error reporting in the GitHub client.
-   **Testing**: Add unit tests for the parser and renderer logic.
-   **Pagination**: optimize the search to handle extremely large contribution histories efficiently.
-   **Configuration**: Add support for a config file (e.g., `.footprint.yml`) to customize aggregation rules and ignore lists.
