package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// TopologyIntelligence synthesizes the structural shape of the candidate's
// portfolio: how concentrated or distributed their work is across ownership,
// maintenance, and originality. It is genuine synthesis — a deterministic
// interpretation of portfolio-level facts — not a re-listing of observations.
//
// Derived from the canonical RepositoryFacts aggregate (never raw observations):
// it answers "is the work concentrated, distributed, or fragmented?" without
// introducing information Atlas does not already own.
type TopologyIntelligence struct {
	dimensionBase
	Level            Level
	OriginalityRatio float64
	TopUpstreamShare float64
	MaintainedShare  float64
	Concentration    string
	Confidence       Confidence
}

// buildTopology synthesizes portfolio topology from the canonical facts
// aggregate. It never reads RepositoryVestige directly; the facts layer is the
// single owner of portfolio aggregation.
func buildTopology(repos []repositoryintelligence.RepositoryIntelligence, f *facts.RepositoryFacts) TopologyIntelligence {
	dim := TopologyIntelligence{}
	dim.name = "topology"

	if f == nil || f.TotalRepos == 0 {
		dim.Level = LevelUnknown
		dim.Concentration = "none"
		dim.Confidence = ConfidenceLow
		dim.summary = "Topology is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("topology", factItem("repositories", 0)),
		}
		return dim
	}

	dim.OriginalityRatio = f.OriginalReposRatio()
	dim.MaintainedShare = f.MaintenanceShare()

	// Ownership concentration: share of forks attributable to the single most
	// forked upstream. 0 when there are no forks.
	dim.TopUpstreamShare = topUpstreamShare(f.ForkLineage)

	// Concentration classification from originality + maintenance distribution.
	dim.Concentration = classifyConcentration(dim.OriginalityRatio, dim.MaintainedShare, dim.TopUpstreamShare)
	dim.Level = topologyLevel(dim.Concentration)

	dim.Confidence = confidenceForSample(f.TotalRepos)
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "topology", []string{"governance", "maintenance", "health"},
			factItem("original repositories", f.OriginalRepos),
			factItem("repositories", f.TotalRepos),
			factItem("originality ratio", fmt.Sprintf("%.2f", dim.OriginalityRatio)),
			factItem("maintained share", fmt.Sprintf("%.2f", dim.MaintainedShare)),
			factItem("top upstream share", fmt.Sprintf("%.2f", dim.TopUpstreamShare)),
			factItem("concentration", dim.Concentration),
		),
	}
	dim.summary = fmt.Sprintf("Topology is %s (%s work; %d%% original, %d%% maintained).", dim.Level, dim.Concentration, pct(dim.OriginalityRatio), pct(dim.MaintainedShare))
	return dim
}

// classifyConcentration expresses the portfolio shape in plain language. It is a
// deterministic function of three portfolio facts; no judgment is introduced.
func classifyConcentration(originality, maintained, topUpstream float64) string {
	switch {
	case originality >= 0.8 && maintained >= 0.6:
		return "concentrated"
	case originality < 0.5:
		return "fork-heavy"
	case topUpstream >= 0.5:
		return "upstream-concentrated"
	case maintained < 0.3:
		return "abandoned-leaning"
	default:
		return "distributed"
	}
}

func topologyLevel(concentration string) Level {
	switch concentration {
	case "concentrated", "distributed":
		return LevelHigh
	case "fork-heavy", "upstream-concentrated":
		return LevelModerate
	case "abandoned-leaning":
		return LevelLow
	default:
		return LevelModerate
	}
}

// topUpstreamShare returns the fraction of forked repositories that descend from
// the single most-forked upstream. 0 when there are no forks.
func topUpstreamShare(lineage []facts.ForkParentFact) float64 {
	if len(lineage) == 0 {
		return 0
	}
	var total, top int
	for _, p := range lineage {
		total += p.Forks
		if p.Forks > top {
			top = p.Forks
		}
	}
	if total == 0 {
		return 0
	}
	return float64(top) / float64(total)
}
