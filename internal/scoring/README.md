# Scoring Algorithm

Each contribution's impact score is calculated based on two metrics:
- Base Score
- Repo popularity

Each activity has a different base score:

|   Activity   | Base Score |
|--------------|------------|
| Pull Request |    10      |
| PR Review    |     3      |
| Issue        |     5      |
| Comment      |     2      |

### Merged Bonus
Merged PRs receive a **1.5x base-score bonus** before the popularity multiplier is applied. This prioritizes accepted contributions.

### Repo Popularity Multiplier
The impact score is adjusted by the repository's adoption and popularity:

```text
impact_score = (base_score * bonus) * popularity_multiplier

popularity_multiplier = 1 + log10(1 + repo_stars + 2*repo_forks)
```

The log multiplier ensures that the impact score increases sublinearly with the number of stars and forks. This ensures that while contribution to popular projects is rewarded, it doesn't disproportionately dwarf other meaningful work.

**Why 2*forks?** Forks signal high-intent adoption and are generally rarer than stars, so they are weighted more heavily to ensure they matter noticeably.

### Repo-Level Aggregation
For ranking "Top Repositories" on the Footprint card, contributions are grouped by repository. The **Total Impact Score** for a repository is the sum of all individual contribution scores made to that project.
