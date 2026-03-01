# Scoring Algorithm

Each contribution's impact score is calculated based on two metrics:
- Base Score
- Repo popularity

Each activity has a different base score:

|   Activity   | Base Score |
|--------------|------------|
| Pull Request |    10      |
| Code Review  |     3      |
| Issue        |     5      |
| Comment      |     2      |

### Merged Bonus
Merged PRs receive a **1.5x base-score bonus** before the popularity multiplier is applied. This prioritizes accepted contributions.

### Repo Popularity Multiplier
The impact score is adjusted by the repository's adoption and popularity:

```text
impact_score = (base_score * bonus) * popularity_multiplier * decay_factor
```

Where `decay_factor` is applied to specific repetitive contribution types (see below).

The `popularity_multiplier` is calculated as:
```text
popularity_multiplier = 1 + log10(1 + repo_stars + 2*repo_forks)
```

The log multiplier ensures that the impact score increases sublinearly with the number of stars and forks. This ensures that while contribution to popular projects is rewarded, it doesn't disproportionately dwarf other meaningful work.

**Why 2*forks?** Forks signal high-intent adoption and are generally rarer than stars, so they are weighted more heavily to ensure they matter noticeably.

### Diminishing Returns (Decay)
To encourage diverse engagement and prevent volume-based gaming of scores per repository, repetitive contributions like comments and issues are subject to a decay factor:

```text
decay_factor = 1.0 / (1.0 + 0.5 * count)
```

Where `count` is the number of existing contributions of a decayable type in that repository. This scales scores as: 1.0x, 0.66x, 0.5x, 0.4x, etc.

**Decayable Types:**
- `IssueComment`
- `ReviewComment`
- `PRComment`
- `DiscussionComment`

### Repo-Level Aggregation
For ranking "Top Repositories" on the Footprint card, contributions are grouped by repository. The **Total Impact Score** for a repository is the sum of all individual contribution scores made to that project.
