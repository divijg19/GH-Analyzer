// This file provides deprecated compatibility aliases for the legacy ontology
// layout. Canonical ownership has moved to internal/observations,
// internal/facts, internal/indicators, and internal/evidence. New code should
// import those packages directly.
package signals

import (
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/indicators"
	"github.com/divijg19/Atlas/internal/observations"
)

// Deprecated: use internal/observations.RepositoryVestige.
type RepositoryVestige = observations.RepositoryVestige

// Deprecated: use internal/facts.RepositoryFacts.
type RepositoryFacts = facts.RepositoryFacts

// Deprecated: use internal/facts.ActivityFacts.
type ActivityFacts = facts.ActivityFacts

// Deprecated: use internal/indicators.Signals.
type Signals = indicators.Signals

// Deprecated: use internal/indicators.RawScore.
type RawScore = indicators.RawScore

// Deprecated: use internal/observations.ActivityObservation.
type ActivityObservation = observations.ActivityObservation

// Deprecated: use internal/observations.ActivityKind.
type ActivityKind = observations.ActivityKind

// Deprecated: use internal/observations.ActivityMetadata.
type ActivityMetadata = observations.ActivityMetadata

// Deprecated: use internal/evidence.Evidence.
type Evidence = evidence.Evidence

// Deprecated: use internal/evidence.EvidenceGroup.
type EvidenceGroup = evidence.EvidenceGroup

const (
	SignalOwnership   = indicators.SignalOwnership
	SignalConsistency = indicators.SignalConsistency
	SignalDepth       = indicators.SignalDepth
	SignalActivity    = indicators.SignalActivity

	ActivityKindCommit             = observations.ActivityKindCommit
	ActivityKindPullRequest        = observations.ActivityKindPullRequest
	ActivityKindReview             = observations.ActivityKindReview
	ActivityKindIssue              = observations.ActivityKindIssue
	ActivityKindDiscussion         = observations.ActivityKindDiscussion
	ActivityKindRelease            = observations.ActivityKindRelease
	ActivityKindActiveDay          = observations.ActivityKindActiveDay
	ActivityKindContributionByRepo = observations.ActivityKindContributionByRepo
	ActivityKindAggregate          = observations.ActivityKindAggregate
)

// Deprecated: use internal/facts.FromRepos.
func FromRepos(repos []RepositoryVestige, referenceTime time.Time) RepositoryFacts {
	return facts.FromRepos(repos, referenceTime)
}

// Deprecated: use internal/facts.ActivityFactsFromObservations.
func ActivityFactsFromObservations(observationsSlice []ActivityObservation, referenceTime time.Time) ActivityFacts {
	return facts.ActivityFactsFromObservations(observationsSlice, referenceTime)
}

// Deprecated: use internal/indicators.ExtractSignals.
func ExtractSignals(repos []RepositoryVestige, referenceTime time.Time) Signals {
	return indicators.ExtractSignals(repos, referenceTime)
}

// Deprecated: use internal/indicators.ExtractSignalsFromFacts.
func ExtractSignalsFromFacts(f RepositoryFacts, referenceTime time.Time) Signals {
	return indicators.ExtractSignalsFromFacts(f, referenceTime)
}

// Deprecated: use internal/indicators.ScoreSignals.
func ScoreSignals(signals Signals) RawScore {
	return indicators.ScoreSignals(signals)
}

// Deprecated: use internal/indicators.SignalsToMap.
func SignalsToMap(s Signals) map[string]float64 {
	return indicators.SignalsToMap(s)
}

// Deprecated: use internal/indicators.FromMap.
func FromMap(m map[string]float64) Signals {
	return indicators.FromMap(m)
}

// Deprecated: use internal/evidence.GenerateEvidence.
func GenerateEvidence(factsValue RepositoryFacts, s Signals) []EvidenceGroup {
	return evidence.GenerateEvidence(factsValue, s)
}
