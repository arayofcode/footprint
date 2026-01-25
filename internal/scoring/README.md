# Scoring Algorithm

Each contribution's impact score is calculated based on two metrics:
- Base Score
- Repo popularity

Each activity has a different base score, which is listed below:
Merged PRs receive a 1.5x base-score bonus before applying repo popularity.
| Activity | Score |
|-|-|
| Pull Request | 10 |
| PR Review | 3 |
| Issue | 5 |
| Comment | 2 |

Each activity can have different attributes, like PR raised v/s merged, comment with high number of upvotes/ downvotes, or comment referenced in PR. These are not being used to determine impact of contribution right now, but will be used in a future iteration.

The impact score also considers the repo popularity

Here's the current scoring algorithm:

```
impact_score = base_score * popularity_multiplier

popularity_multiplier = 1 + log10(1 + repo_stars + 2*repo_forks)
```

The log multiplier ensures that the impact score increases sublinearly with the number of stars and forks, and ensures a small contribution in a popular repo doesn't get disproportionately high impact score.

IMO forks signal repo's adoption, and the number of forks is generally significantly lesser than the stars. I want forks to matter noticeably, but not overpower the stars in the repo popularity multiplier, hence the 2*forks.

Later, the algorithm will also consider a reaction multiplier, which takes into account:
- Comment leading to issue resolution or PR
- Comment tagged referred in another issue or PR
- Number of contributions made to a repo where user is an active contributor but not member of core team
- Number of reactions to a commit or PR (emoji, comments or reviews).
    - This can be tricky as someone's spammy PR might have high number of reactions.
