package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// MaintenanceIntelligence quantifies how actively the portfolio is maintained.
// Aggregated from per-repository Repository Intelligence.
type MaintenanceIntelligence struct {
	dimensionBase
	Level         Level
	MaintainedRatio float64
	StaleRepos    int
	ArchivedRepos int
	Confidence    Confidence
}

func buildMaintenance(repos []repositoryintelligence.RepositoryIntelligence) MaintenanceIntelligence {
	dim := MaintenanceIntelligence{}
	dim.name = "maintenance"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Maintenance is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("maintenance", factItem("repositories", 0)),
		}
		return dim
	}

	var maintained, stale, archived int
	for _, r := range repos {
		if r.Maintenance.Level == repositoryintelligence.LevelMaintained {
			maintained++
		}
		if r.Maintenance.Level == repositoryintelligence.LevelInactive {
			stale++
		}
		if r.Maintenance.Archived {
			archived++
		}
	}

	ratio := float64(maintained) / float64(total)
	dim.MaintainedRatio = ratio
	dim.StaleRepos = stale
	dim.ArchivedRepos = archived
	dim.Level = levelFromRatio(ratio, 0.7, 0.4)
	dim.Confidence = ConfidenceHigh
	dim.evidence = []evidence.EvidenceGroup{
		group("maintenance",
			factItem("maintained repositories", maintained),
			factItem("repositories", total),
			factItem("stale repositories", stale),
			factItem("archived repositories", archived),
		),
	}
	dim.summary = fmt.Sprintf("Maintenance is %s (%d of %d repositories actively maintained).", dim.Level, maintained, total)
	return dim
}
