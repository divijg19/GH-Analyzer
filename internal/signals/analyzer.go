// Package signals owns the Vestiges, Facts, and Signals layers of the Atlas
// Intelligence Ontology (see docs/INTELLIGENCE.md). It owns:
//
//   - RepositoryVestige — the canonical observation of a repository
//   - RepositoryFacts — deterministic aggregates derived from vestiges
//   - Signals — normalized measurements (Ownership, Consistency, Depth, Activity)
//   - RawScore — unscored evaluation inputs
//
// These three layers are co-located because they share the same change
// drivers: new acquisition backends produce new vestige fields, which flow
// into new fact aggregations, which may drive new signals. Splitting them
// would add package boundaries without improving isolation.
//
// This package owns no overall scoring, confidence, or ranking — those
// belong to internal/evaluation.
package signals

import (
	"math"
	"time"
)

const (
	recentWindowDays  = 90
	minDepthRepoSize  = 50
	minConsistencyDen = 10
	minDepthDen       = 5

	activityHighDays     = 30
	activityModerateDays = 90
	activityLowDays      = 180

	activityHighValue     = 1.0
	activityModerateValue = 0.7
	activityLowValue      = 0.4
	activityStaleValue    = 0.1

	scoreScale     = 100.0
	minSignalValue = 0.0
	maxSignalValue = 1.0
)

// RepositoryVestige is the canonical observation of a repository. Every
// acquisition backend (REST, GraphQL, GitFut, local git) produces vestiges;
// Atlas never acquires from them directly. Fields are grouped into observation
// domains: Identity, Ownership, Timeline, Technology, Maintenance, Structure.
//
// RepositoryVestige represents every observation Atlas currently understands.
// It does not mirror provider schemas. Growth is deliberate: every new
// observation must justify its existence.
type RepositoryVestige struct {
	// Identity
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	Archived   bool   `json:"archived"`
	Template   bool   `json:"template"`

	// Ownership
	Fork              bool   `json:"fork"`
	ParentRepository  string `json:"parent_repository,omitempty"`
	CollaboratorCount int    `json:"collaborator_count"`

	// Timeline
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	LatestReleaseAt time.Time `json:"latest_release_at,omitempty"`
	ReleaseCount    int       `json:"release_count"`

	// Technology
	License                string           `json:"license"`
	Topics                 []string         `json:"topics"`
	DefaultBranch          string           `json:"default_branch"`
	DefaultBranchProtected bool             `json:"default_branch_protected"`
	LanguageDistribution   map[string]int64 `json:"language_distribution,omitempty"`

	// Maintenance
	OpenIssues       int `json:"open_issues_count"`
	Stars            int `json:"stargazers_count"`
	Forks            int `json:"forks_count"`
	Watchers         int `json:"watchers_count"`
	PullRequestCount int `json:"pull_request_count"`

	// Structure
	Size              int  `json:"size"`
	DiscussionEnabled bool `json:"discussion_enabled"`
}

type Signals struct {
	Ownership   float64
	Consistency float64
	Depth       float64
	Activity    float64
}

type RawScore struct {
	Ownership   int `json:"ownership"`
	Consistency int `json:"consistency"`
	Depth       int `json:"depth"`
}

func ExtractSignals(repos []RepositoryVestige, referenceTime time.Time) Signals {
	return ExtractSignalsFromFacts(FromRepos(repos, referenceTime), referenceTime)
}

func ExtractSignalsFromFacts(f RepositoryFacts, referenceTime time.Time) Signals {
	consistency := 0.0
	depth := 0.0
	ownership := 0.0

	if f.OriginalRepos > 0 {
		consistencyDen := maxInt(minConsistencyDen, f.OriginalRepos)
		depthDen := maxInt(minDepthDen, f.OriginalRepos)
		consistency = float64(f.RecentRepos) / float64(consistencyDen)
		depth = float64(f.DeepRepos) / float64(depthDen)
	}

	if f.ValidRepos > 0 {
		ownership = float64(f.ValidOriginalRepos) / float64(f.ValidRepos)
	}

	activity := activityFromTime(f.LatestActivity, referenceTime)

	return Signals{
		Ownership:   clamp01(ownership),
		Consistency: clamp01(consistency),
		Depth:       clamp01(depth),
		Activity:    clamp01(activity),
	}
}

func ScoreSignals(signals Signals) RawScore {
	ownership := signalToScore(signals.Ownership)
	consistency := signalToScore(signals.Consistency)
	depth := signalToScore(signals.Depth)

	return RawScore{
		Ownership:   ownership,
		Consistency: consistency,
		Depth:       depth,
	}
}

func activityFromTime(latest time.Time, now time.Time) float64 {
	if latest.IsZero() {
		return 0
	}

	daysSinceLastUpdate := int(now.Sub(latest).Hours() / 24)

	switch {
	case daysSinceLastUpdate <= activityHighDays:
		return activityHighValue
	case daysSinceLastUpdate <= activityModerateDays:
		return activityModerateValue
	case daysSinceLastUpdate <= activityLowDays:
		return activityLowValue
	default:
		return activityStaleValue
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func signalToScore(value float64) int {
	if value < minSignalValue {
		value = minSignalValue
	}
	if value > maxSignalValue {
		value = maxSignalValue
	}

	scaled := value * scoreScale
	return int(math.Round(scaled))
}
