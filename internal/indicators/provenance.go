package indicators

import (
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/provenance"
)

// signalFacts maps each signal to the repository facts it consumes. Indicators
// are backend-agnostic: they know only facts, never observations. The mapping is
// static because ExtractSignalsFromFacts is a fixed, pure derivation.
var signalFacts = map[string][]string{
	SignalOwnership:   {"valid_original_repos", "valid_repos"},
	SignalConsistency: {"recent_repos", "original_repos"},
	SignalDepth:       {"deep_repos", "original_repos", "largest_repo_size"},
	SignalActivity:    {"latest_activity"},
}

// SignalProvenance returns the complete deterministic provenance chain for a
// signal: the indicator itself, the facts it consumed, and — by transitively
// resolving those facts — the repository observation fields underneath. This
// makes each signal explainable down to the observations without any layer
// duplicating another's knowledge.
func SignalProvenance(signal string) provenance.Chain {
	factNames := signalFacts[signal]
	chain := provenance.Chain{Indicators: provenance.Indicators(signal)}
	return provenance.Merge(chain, facts.FactsProvenance(factNames...))
}
