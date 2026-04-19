# GitHub Signal Analyzer

GH-Analyzer is a minimal Go CLI that analyzes a GitHub user's public repositories and returns a deterministic JSON report.

## Project Scope

- CLI: deterministic analysis, scoring, search/query workflows, and JSON/report output.
- Frontend UI: lightweight visualization and exploration of analyzer results without changing scoring logic.

## Usage

- gh-analyzer <username>
- gha <username>

If username is missing, the CLI exits with an error message.

## Signal Definitions

- Consistency: recent non-fork repositories divided by max(10, total non-fork repositories); measures recent activity across original work.
- Depth: deep non-fork repositories (size >= 50) divided by max(5, total non-fork repositories); measures substantial original projects.
- Ownership: non-fork repositories with size > 0 divided by all repositories with size > 0; ignores trivial size-0 repositories.
- Activity: based on the latest update across all repositories with graded decay:
  - <= 30 days: 1.0
  - <= 90 days: 0.7
  - <= 180 days: 0.4
  - over 180 days: 0.1

## Scoring

Signals are clamped to [0,1] and converted to 0-100. Overall score uses:

- Consistency: 40%
- Ownership: 30%
- Depth: 30%

For very small datasets (fewer than 3 repos), overall score is multiplied by 0.7.

## Scoring Interpretation

Scores are percentile-based within the dataset.

A score of 0.90 means:
	the candidate ranks higher than 90% of the dataset.

## Live Mode

Use --live to fetch candidates directly from GitHub.

Example:
	gh-analyzer search backend --live

Notes:
- limited sample size (max ~20 users)
- subject to GitHub API rate limits
- results are not persisted unless explicitly saved

## Example Output (JSON)

~~~json
{
	"username": "octocat",
	"scores": {
		"Ownership": 75,
		"Consistency": 100,
		"Depth": 100,
		"Overall": 93
	},
	"summary": "Strong consistency, strong ownership, strong depth",
	"highlights": [
		"Active in last 90 days",
		"Majority original repositories",
		"Includes non-trivial projects"
	]
}
~~~

## Limitations

- Uses only public repositories.
- No authentication, so GitHub API rate limits apply.
- Repository size is a coarse proxy for project depth.
- Results depend on current GitHub metadata and time window.
