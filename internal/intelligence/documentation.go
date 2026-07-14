package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// DocumentationIntelligence quantifies how well the portfolio is documented.
// Aggregated from the per-repository documentation dimension of Repository
// Intelligence.
type DocumentationIntelligence struct {
	dimensionBase
	Level           Level
	DocumentedRatio float64
	DocumentedRepos int
	LicensedRepos   int
	Confidence      Confidence
}

func buildDocumentation(repos []repositoryintelligence.RepositoryIntelligence) DocumentationIntelligence {
	dim := DocumentationIntelligence{}
	dim.name = "documentation"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Documentation is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("documentation", factItem("repositories", 0)),
		}
		return dim
	}

	var documented, licensed int
	for _, r := range repos {
		switch r.Documentation.Level {
		case repositoryintelligence.LevelHigh, repositoryintelligence.LevelModerate:
			documented++
		}
		if r.Documentation.HasLicense {
			licensed++
		}
	}

	ratio := float64(documented) / float64(total)
	dim.DocumentedRatio = ratio
	dim.DocumentedRepos = documented
	dim.LicensedRepos = licensed
	dim.Level = levelFromRatio(ratio, 0.6, 0.3)
	dim.Confidence = ConfidenceHigh
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "documentation", []string{"documentation"},
			factItem("documented repositories", documented),
			factItem("repositories", total),
			factItem("licensed repositories", licensed),
		),
	}
	dim.summary = fmt.Sprintf("Documentation is %s (%d of %d repositories are documented).", dim.Level, documented, total)
	return dim
}
