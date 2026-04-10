# GitHub Signal Analyzer

GH-Analyzer is a minimal Go CLI that analyzes a GitHub user's public repositories and outputs a structured JSON report.

## Problem

Given a GitHub username, quickly estimate repository activity and quality signals without external dependencies or complex infrastructure.

## Approach

1. Accept a GitHub username as a CLI argument.
2. Fetch public repositories from the GitHub REST API.
3. Compute signal values from repository metadata.
4. Convert signals into weighted scores.
5. Return a final JSON report with scores, summary, and highlights.

Use GH-Analyzer through either CLI entrypoint:

- gh-analyzer <username>
- gha <username>

If username is missing, the CLI exits with an error message.

## Signals

- Consistency: ratio of repositories updated in the last 90 days.
- Ownership: ratio of non-fork repositories.
- Depth: average size of non-fork repositories where size is at least 50.

Each signal is converted to a 0-100 score. Overall score uses weighted aggregation:

- Consistency: 40%
- Ownership: 30%
- Depth: 30%

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

## Future Work

- Add optional authenticated requests for higher rate limits.
- Add organization-level analysis.
- Add additional signals (language diversity, contribution cadence).
- Add optional output modes for machine pipelines.
