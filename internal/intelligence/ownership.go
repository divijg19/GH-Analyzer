package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// OwnershipIntelligence quantifies how much of the portfolio is original work
// rather than forks. Aggregated from per-repository Repository Intelligence.
type OwnershipIntelligence struct {
	dimensionBase
	Level            Level
	OriginalityRatio float64
	ForkRatio        float64
	Collaborators    int
	Confidence       Confidence
}

func buildOwnership(repos []repositoryintelligence.RepositoryIntelligence) OwnershipIntelligence {
	dim := OwnershipIntelligence{}
	dim.name = "ownership"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Ownership is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("ownership", factItem("repositories", 0)),
		}
		return dim
	}

	var original, forks, collaborators int
	for _, r := range repos {
		if !r.Governance.Fork {
			original++
		} else {
			forks++
		}
		collaborators += r.Governance.Collaborators
	}

	ratio := float64(original) / float64(total)
	dim.OriginalityRatio = ratio
	dim.ForkRatio = float64(forks) / float64(total)
	dim.Collaborators = collaborators
	dim.Level = levelFromRatio(ratio, 0.7, 0.4)
	dim.Confidence = ConfidenceHigh
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "ownership", []string{"governance"},
			factItem("original repositories", original),
			factItem("repositories", total),
			factItem("fork repositories", forks),
			factItem("total collaborators", collaborators),
		),
	}
	dim.summary = fmt.Sprintf("Ownership is %s (%d%% of %d repositories are original).", dim.Level, pct(ratio), total)
	return dim
}
