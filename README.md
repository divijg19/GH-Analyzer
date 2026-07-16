# `Atlas`

Atlas analyzes GitHub users' public repositories and returns deterministic signal-based profiles, enabling search, comparison, and evaluation.

## Project Scope

- **CLI**: analysis, search, query, inspect, dataset management
- **Server**: HTTP API for search and analysis
- **Frontend**: lightweight visualization (see `web/`)
- **Pure pipeline**: Acquisition → Normalization → Vestiges → Facts → Indicators → Profile → Evaluation → Projection → Consumers

## Commands

- `atlas <username>` — Analyze a single GitHub profile
- `atlas search <query>` — Discover developers by intent or expression
- `atlas query [options]` — Advanced signal-threshold query
- `atlas inspect <username>` — Inspect a profile in a dataset
- `atlas build <usernames>` — Build a dataset from usernames
- `atlas dataset [info|preview|stats]` — Show dataset summary
- `atlas serve` — Start the HTTP API server

## Signal Definitions

- **Consistency**: recent non-fork repositories divided by max(10, total non-fork repositories); measures recent activity across original work.
- **Depth**: deep non-fork repositories (size >= 50 KB) divided by max(5, total non-fork repositories); measures substantial original projects.
- **Ownership**: non-fork repositories with size > 0 divided by all repositories with size > 0; ignores trivial size-0 repositories.
- **Activity**: based on the latest update across all repositories with graded decay:
  - <= 30 days: 1.0
  - <= 90 days: 0.7
  - <= 180 days: 0.4
  - over 180 days: 0.1

## Scoring

Signals are clamped to [0,1] and converted to 0-100 integer component scores. Overall score uses:

- Consistency: 40%
- Ownership: 30%
- Depth: 30%

For very small datasets (fewer than 3 repos), overall score is multiplied by 0.7.

Within a dataset of 10 or more profiles, the final search score is
percentile-calibrated; smaller datasets use the raw weighted score.

## Architecture

```
GitHub
   ↓
Transport          internal/github
   ↓
Acquisition        internal/acquisition
   ↓
Normalization      internal/acquisition/normalize.go
   ↓
Vestiges           signals.Repo · profile.UserMetadata · contributions.Summary
   ↓
Facts              signals.Facts
   ↓
Indicators         signals.Signals → RawScore
   ↓
Profile            index.Profile
   ↓
Evaluation         internal/evaluation (overall score, penalty, confidence, ranking policy)
   ↓
Intelligence       Profile + Evaluation output
   ↓
Projection         internal/projection
   ↓
Consumers          cmd/atlas · cmd/server · web
```

Search execution (`internal/engine`) sits between Evaluation and Projection,
filtering and ranking profiles using the evaluation policy.

Each layer has strict ownership. See `docs/ARCHITECTURE.md` for layer
ownership and `docs/INTELLIGENCE.md` for the canonical intelligence model.

## Search Modes

Use `--live` to fetch candidates directly from GitHub:

```
atlas search backend --live
```

Limited to ~20 users, subject to GitHub API rate limits.

## Limitations

- Uses only public repositories.
- No authentication, so GitHub API rate limits apply.
- Repository size is a coarse proxy for project depth.
- Results depend on current GitHub metadata and time window.
