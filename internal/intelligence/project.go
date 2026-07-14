package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// ProjectIntelligence quantifies the scale and maturity distribution of the
// portfolio's repositories. Aggregated from per-repository Repository
// Intelligence.
type ProjectIntelligence struct {
	dimensionBase
	Level         Level
	TotalSize     int
	MatureRepos   int
	EarlyRepos    int
	DistinctStacks int
	Confidence    Confidence
}

func buildProject(repos []repositoryintelligence.RepositoryIntelligence) ProjectIntelligence {
	dim := ProjectIntelligence{}
	dim.name = "project"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Project is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("project", factItem("repositories", 0)),
		}
		return dim
	}

	var totalSize, mature, early int
	stackSet := make(map[string]bool)
	for _, r := range repos {
		totalSize += r.Architecture.Size
		if r.Lifecycle.Level == repositoryintelligence.LevelMature {
			mature++
		}
		if r.Lifecycle.Level == repositoryintelligence.LevelEarly {
			early++
		}
		if r.Technology.PrimaryLang != "" {
			stackSet[r.Technology.PrimaryLang] = true
		}
	}

	dim.TotalSize = totalSize
	dim.MatureRepos = mature
	dim.EarlyRepos = early
	dim.DistinctStacks = len(stackSet)

	switch {
	case total > 0 && mature >= total/2:
		dim.Level = LevelBroad
	case total > 0 && early >= total/2:
		dim.Level = LevelNarrow
	default:
		dim.Level = LevelModerate
	}
	dim.Confidence = ConfidenceHigh
	dim.evidence = []evidence.EvidenceGroup{
		group("project",
			factItem("total size (KB)", totalSize),
			factItem("mature repositories", mature),
			factItem("early repositories", early),
			factItem("distinct primary languages", dim.DistinctStacks),
		),
	}
	dim.summary = fmt.Sprintf("Project is %s (%d mature of %d repositories; %dKB total).", dim.Level, mature, total, totalSize)
	return dim
}
