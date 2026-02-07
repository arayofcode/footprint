# Footprint

**Tagline:** *OSS contributions you forgot about*

## 1) One-liner

Footprint is a GitHub Action + embeddable SVG that discovers a user’s public open‑source contributions across GitHub (beyond just PRs) and generates a portfolio-ready “footprint” card + report.

---

## 2) Problem

GitHub profiles are good at showing repo ownership, commit graphs, and PR counts, but are bad at surfacing:

* cross-repo contributions outside your own repositories
* reviews, comments, and discussions that helped others ship
* older contributions people no longer remember

As a result, many contributors (especially early-career engineers) have real OSS impact that is hard to showcase.

---

## 3) Goal

For a given GitHub username, automatically generate:

1. an **embeddable SVG card** for GitHub README / personal website
2. a **JSON report** with all discovered contribution events
3. a human-readable **summary.md**

The output should be easy to publish without running a server.

---

## 4) Why GitHub Action (snk-style)

Footprint is designed to work like `Platane/snk`:

* runs on GitHub Actions (no server cost)
* outputs files to a separate branch (e.g., `output`)
* SVG is embedded from that branch in README / website

This makes setup extremely simple:

* copy workflow YAML
* automatically uses `GITHUB_ACTOR` (no username config needed)
* run the action
* embed the generated SVG

> Note: For hosted usage (e.g., an EC2 deployment), the username is a request parameter instead of `GITHUB_ACTOR`.

---

## 5) Scope

Footprint currently focuses on two buckets:

#### A) **Owned OSS Projects ("Project Score")**

These are repositories that the user owns and that meet a minimal popularity threshold.

* repo owner is the user
* repo is public
* repo is not a fork
* **stars ≥ 5** (configurable)

**Owned projects are *not* treated as contribution events** to avoid self-inflation. The card displays these with a clean layout focused on repo name and star count.

* **External Contributions ("Contribution Score")**
    * Activities in repositories the user does **not** own.
    * **Calculated Scores**: PRs authored, Code Review (Reviews/Comments), Issues opened, and Comments.
    * **Merged detection**: Merged Pull Requests receive a **1.5x score boost**.
    * **Top External Impact**: Grouped and ranked by **Total Impact Score**.

---

### Planned Features
* Wiki contributions
* Reaction-based impact indexing (reactions, references, answer markers)

---

### Output artifacts

* `dist/report.json`
* `dist/summary.md`
* `dist/card.svg`

### CLI usage (local dev)

Run either entrypoint:

* `go run ./cmd/footprint`
* `go run .`

Optional flags:

* `-username` (defaults to `GITHUB_ACTOR`)
* `-min-stars` (owned project threshold)
* `-output` (output directory)
* `-clamp` (popularity multiplier cap)
* `-timeout` (GitHub API timeout)
* `-card` (enable/disable SVG card output)

Environment:

* `GITHUB_TOKEN` (required)
* `GITHUB_ACTOR` (used when `-username` is not set)

### Ranking (no LLM required)

Use heuristics to compute an impact score per event.

---

## 6) Out of scope (initially)

* Private repos/org contributions
* Perfect attribution chains (comment → patch implemented)
* LLM-based summarization (optional later)
* Multi-platform support (GitLab, Bitbucket)

---

## 7) User story

As a developer who contributes across GitHub, I want to generate a “Footprint” report that surfaces contributions I forgot about and lets me embed the highlights into my README/website.

---

## 8) Functional requirements

1. **Input:** GitHub username (+ optional time range)
2. Fetch contribution events from GitHub APIs
3. Normalize into a common `ContributionEvent` format
4. Compute an impact score (heuristic)
5. Generate report artifacts (JSON/MD/SVG)
    * **SVG Card**:
        * **Modular Design**: Compact "card-within-a-card" rows with rounded corners.
        * **Interactive Layers**: Entire rows are clickable; individual metrics link to specialized GitHub search queries.
        * **Grid Alignment**: Standardized metric columns (normalized spacing) for a balanced, professional grid.
        * **Intelligent Avatars**: Full-circle clipped avatars for users and repositories.
6. Publish artifacts to an output branch
7. Provide embed snippet for README/website

---

## 9) Non-functional requirements

* No external server required (GitHub Action mode)
* GraphQL-first: minimizes API calls by fetching repo stats and PR details in one query tree where possible
* Rate-limit aware (uses `GITHUB_TOKEN` for ~5000/hr limit)
* Deterministic outputs for same inputs (except new data)
* SVG renders correctly in GitHub README and browsers
* Setup should take < 5 minutes

---

## 10) Data model

### ContributionEvent (external activity)

* `id` (stable; recommended: `<owner>/<repo>#<number>` for PRs)
* `type` (PR | ISSUE | ISSUE_COMMENT | PR_FEEDBACK | PR_COMMENT | DISCUSSION | DISCUSSION_COMMENT)
* `repo` (owner/name)
* `url`
* `title` (if applicable)
* `createdAt`
* `stars` (repo stargazer count)
* `forks` (repo fork count)
* `merged` (bool, for PRs)
* `answer` (bool, for discussions)
* `reactionsCount` (total reactions count)
* `score` (computed locally)

### OwnedProject (ownership bucket)

* `repo` (owner/name)
* `url` (repository link)
* `stars` (repo stargazer count)
* `forks` (repo fork count)
* `score` (computed locally)
* **Sorting**: Ranked by **Stars (desc)** then **Forks (desc)**.

---

## 11) Scoring algorithm

Each contribution’s impact score is calculated based on two metrics:

* **Base Score** (depends on activity type)
* **Repo Popularity** (stars + forks)

### Definitions

* `repo_stars`: repository **stargazer count**
* `repo_forks`: repository **fork count**
* **Comment**: issue comment or PR comment (future: discussion comments)

### Base Score
Each activity has a different base score:

| Pull Request      |          10 |
| PR Review         |           5 |
| PR Comment        |           2 |
| Issue             |           5 |
| Issue Comment     |           2 |

> [!NOTE]
> PR Review specifically refers to a submitted review (Approve/Request Changes/Comment), while PR Comment (Feedback) refers to individual file-level comments or general comments on a PR.

### Scoring Attributes
Each activity can have several attributes (e.g., merged PR vs unmerged PR, comment referenced by a PR, etc.). These contribute to the final impact score by applying multipliers or bonuses.

### Repo Popularity Multiplier

```text
impact_score = base_score * popularity_multiplier

popularity_multiplier = 1 + log10(1 + repo_stars + 2*repo_forks)
```

The log multiplier ensures that the impact score increases sublinearly with the number of stars and forks, and ensures a small contribution in a popular repo doesn't get disproportionately high impact score.

**Why `2*forks`:** Forks signal adoption and contributor intent (fork → modify → PR). Fork counts are usually significantly lower than stars, so weighting them higher ensures forks matter noticeably without overpowering stars.

`popularity_multiplier = min(4.0, 1 + log10(1 + repo_stars + 2*repo_forks))`

### Future: Reaction / Reference Multiplier

Later, the algorithm may add a reaction/reference multiplier based on:

* Comment leading to issue resolution or PR
* Comment referenced in another issue or PR
* Number of contributions to a repo where the user is an active contributor but not a core team member
* Number of reactions to a PR/issue/comment/review (emoji reactions)

  * Note: this can be noisy (spammy PRs may receive many reactions), so reaction boosts should be capped and weighted conservatively.

---

## 12) Success criteria (quantified)

For a sample of users:

* ≥95% of runs successfully generate artifacts
* ≥30% of discovered events are non-PR events
* p95 runtime ≤ 60 seconds for typical users
* ≥20% of users copy the embed snippet (if analytics added)

---

## 13) License

Recommended: **Apache-2.0** + `NOTICE` for attribution.

NOTICE should use the legal name:

* **Aryan Sharma** (aka Ray)

---

## 14) Repo structure (suggested)

```text
footprint/
  .github/workflows/footprint.yml
  cmd/footprint/
    main.go
  internal/
    github/
      client.go
      types.go
    scoring/
      scoring.go
    render/
      assets/
        fetcher.go
      card/
        renderer.go
        viewmodel.go
        layout.go
      report.go
      summary.md
  dist/ (generated)
  Dockerfile
  action.yaml
  go.mod
  README.md
  Notes.md
```

---

## 15) Antigravity-specific context

The primary deliverable is a GitHub Action that generates a public SVG card + report for a user’s GitHub OSS activity.
The project should be easy to fork, configure, and embed (similar to other README widgets).
