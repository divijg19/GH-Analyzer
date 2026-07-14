package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// BreadthIntelligence quantifies language and stack diversity across the
// portfolio. Aggregated from per-repository Repository Intelligence.
type BreadthIntelligence struct {
	dimensionBase
	Level           Level
	DistinctLangs   int
	MultiLangRepos  int
	MaxLangInRepo   int
	Confidence      Confidence
}

func buildBreadth(repos []repositoryintelligence.RepositoryIntelligence) BreadthIntelligence {
	dim := BreadthIntelligence{}
	dim.name = "breadth"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Breadth is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("breadth", factItem("repositories", 0)),
		}
		return dim
	}

	langSet := make(map[string]bool)
	var multiLang, maxLang int
	for _, r := range repos {
		if r.Technology.LanguageCount >= 2 {
			multiLang++
		}
		if r.Technology.LanguageCount > maxLang {
			maxLang = r.Technology.LanguageCount
		}
		for _, l := range r.Technology.Stack {
			langSet[l] = true
		}
	}

	dim.DistinctLangs = len(langSet)
	dim.MultiLangRepos = multiLang
	dim.MaxLangInRepo = maxLang

	switch {
	case dim.DistinctLangs >= 5 || (total > 0 && multiLang >= total/2):
		dim.Level = LevelBroad
	case dim.DistinctLangs <= 1:
		dim.Level = LevelNarrow
	default:
		dim.Level = LevelModerate
	}
	dim.Confidence = ConfidenceHigh
	dim.evidence = []evidence.EvidenceGroup{
		group("breadth",
			factItem("distinct languages", dim.DistinctLangs),
			factItem("repositories with 2+ languages", multiLang),
			factItem("max languages in one repository", maxLang),
		),
	}
	dim.summary = fmt.Sprintf("Breadth is %s (%d distinct languages across %d repositories).", dim.Level, dim.DistinctLangs, total)
	return dim
}
