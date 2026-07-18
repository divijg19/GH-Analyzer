package intelligence

import (
	"fmt"
	"sort"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// FocusIntelligence quantifies specialization: how concentrated the portfolio is
// around a dominant technology, as opposed to spread thin. It is the semantic
// complement of Breadth — Breadth measures diversity, Focus measures
// concentration. Aggregated from per-repository Repository Intelligence.
type FocusIntelligence struct {
	dimensionBase
	Level            Level
	DominantLanguage string
	DominantShare    float64
	DistinctLangs    int
	Confidence       Confidence
}

func buildFocus(repos []repositoryintelligence.RepositoryIntelligence) FocusIntelligence {
	dim := FocusIntelligence{}
	dim.name = "focus"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Focus is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("focus", factItem("repositories", 0)),
		}
		return dim
	}

	counts := make(map[string]int)
	var withLang int
	for _, r := range repos {
		lang := r.Technology.PrimaryLang
		if lang == "" {
			continue
		}
		counts[lang]++
		withLang++
	}

	if withLang == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Focus is unknown (no primary languages detected)."
		dim.evidence = []evidence.EvidenceGroup{
			group("focus", factItem("repositories with a primary language", 0)),
		}
		return dim
	}

	// Deterministic dominant language: highest count, then lexical order.
	langs := make([]string, 0, len(counts))
	for l := range counts {
		langs = append(langs, l)
	}
	sort.Slice(langs, func(i, j int) bool {
		if counts[langs[i]] != counts[langs[j]] {
			return counts[langs[i]] > counts[langs[j]]
		}
		return langs[i] < langs[j]
	})

	dominant := langs[0]
	share := float64(counts[dominant]) / float64(withLang)
	dim.DominantLanguage = dominant
	dim.DominantShare = share
	dim.DistinctLangs = len(counts)

	switch {
	case share >= 0.6:
		dim.Level = LevelHigh
	case share >= 0.35:
		dim.Level = LevelModerate
	default:
		dim.Level = LevelLow
	}
	dim.Confidence = confidenceForSample(len(repos))
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "focus", []string{"technology"},
			factItem("dominant language", dominant),
			factItem("dominant share (%)", pct(share)),
			factItem("distinct primary languages", len(counts)),
			factItem("repositories with a primary language", withLang),
		),
	}
	dim.summary = fmt.Sprintf("Focus is %s (%s is the primary language in %d%% of repositories).", dim.Level, dominant, pct(share))
	return dim
}
