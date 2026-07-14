package observations

import "time"

// ActivityKind enumerates the kinds of activity observations.
type ActivityKind string

const (
	ActivityKindCommit             ActivityKind = "commit"
	ActivityKindPullRequest        ActivityKind = "pull_request"
	ActivityKindReview             ActivityKind = "review"
	ActivityKindIssue              ActivityKind = "issue"
	ActivityKindDiscussion         ActivityKind = "discussion"
	ActivityKindRelease            ActivityKind = "release"
	ActivityKindActiveDay          ActivityKind = "active_day"
	ActivityKindContributionByRepo ActivityKind = "contribution_by_repo"
	ActivityKindAggregate          ActivityKind = "aggregate"
)

// ActivityMetadata holds the typed metadata currently required by the activity
// observations Atlas acquires in v0.8.17. It intentionally stays narrow so it
// does not become an unbounded bag of values.
type ActivityMetadata struct {
	WindowStart     time.Time
	WindowEnd       time.Time
	ActiveDays      int
	TotalDays       int
	Year            int
	RepoCommitCount int
	RestrictedCount int
}

// ActivityObservation is the canonical raw activity observation used by the
// v0.8.17 activity pipeline.
type ActivityObservation struct {
	Kind       ActivityKind
	Count      int
	Repository string
	Actor      string
	OccurredAt time.Time
	Metadata   ActivityMetadata
}

// ObservationID returns the deterministic, Atlas-owned identity of this activity
// observation. Identity is constructed from the observation kind, its repository
// (if any) and its occurrence instant, which together uniquely and stably
// identify an activity observation across executions.
func (o ActivityObservation) ObservationID() string {
	id := string(o.Kind)
	if o.Repository != "" {
		id += ":" + o.Repository
	}
	if !o.OccurredAt.IsZero() {
		id += "@" + o.OccurredAt.UTC().Format(time.RFC3339)
	}
	return id
}
