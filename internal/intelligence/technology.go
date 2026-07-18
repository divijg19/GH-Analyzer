package intelligence

import (
	"fmt"
	"sort"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

func sortInts(s []int) {
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
}

// TechnologyIntelligence synthesizes the candidate's technology trajectory from
// the lossless TechnologyTimeline: diversification (languages ever used),
// specialization (dominant-language persistence), and transitions (languages
// appearing or disappearing between creation years). It is deterministic
// synthesis over existing facts; it collects nothing new.
type TechnologyIntelligence struct {
	dimensionBase
	Level           Level
	Diversification int
	Specialization  float64
	Transitions     int
	LatestLanguages []string
	Confidence      Confidence
}

func buildTechnology(repos []repositoryintelligence.RepositoryIntelligence, f *facts.RepositoryFacts) TechnologyIntelligence {
	dim := TechnologyIntelligence{}
	dim.name = "technology"

	if f == nil {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Technology is unknown (no language observations)."
		dim.evidence = []evidence.EvidenceGroup{
			group("technology", factItem("creation years with languages", 0)),
		}
		return dim
	}

	timeline := f.TechnologyTimeline
	if len(timeline) == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Technology is unknown (no language observations)."
		dim.evidence = []evidence.EvidenceGroup{
			groupFrom(repos, "technology", []string{"technology"}, factItem("creation years with languages", 0)),
		}
		return dim
	}

	// Diversification: distinct languages observed across all creation years.
	seen := make(map[string]bool)
	years := make([]int, 0, len(timeline))
	for y, langs := range timeline {
		years = append(years, y)
		for _, l := range langs {
			seen[l] = true
		}
	}
	dim.Diversification = len(seen)

	// Specialization: the most common dominant language across years, expressed
	// as its share of years in which a language appears.
	dim.Specialization, dim.LatestLanguages = technologySpecialization(timeline, years)

	// Transitions: languages appearing in a later year that were absent in all
	// earlier years (adoption), or disappearing after appearing (drop).
	dim.Transitions = technologyTransitions(timeline, years)

	dim.Level = technologyLevel(dim.Diversification, dim.Specialization)
	dim.Confidence = confidenceForSample(len(years))
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "technology", []string{"technology"},
			factItem("creation years with languages", len(years)),
			factItem("distinct languages", dim.Diversification),
			factItem("specialization", fmt.Sprintf("%.2f", dim.Specialization)),
			factItem("technology transitions", dim.Transitions),
			factItem("latest-year languages", joinLang(dim.LatestLanguages)),
		),
	}
	dim.summary = fmt.Sprintf("Technology is %s (%d distinct languages; specialization %.2f; %d transitions).", dim.Level, dim.Diversification, dim.Specialization, dim.Transitions)
	return dim
}

// technologySpecialization returns the share of creation years in which the
// single most persistent language appears, plus the languages of the latest
// creation year (the current technology footprint).
func technologySpecialization(timeline map[int][]string, years []int) (float64, []string) {
	yearLangCount := make(map[string]int)
	for _, langs := range timeline {
		for _, l := range langs {
			yearLangCount[l]++
		}
	}
	top := 0
	for _, c := range yearLangCount {
		if c > top {
			top = c
		}
	}
	specialization := 0.0
	if len(years) > 0 {
		specialization = float64(top) / float64(len(years))
	}

	sortInts(years)
	latest := years[len(years)-1]
	latestLangs := append([]string(nil), timeline[latest]...)
	sort.Strings(latestLangs)
	return specialization, latestLangs
}

// technologyTransitions counts deterministic adoptions (a language first seen in
// a later year) and drops (a language absent from any later year after earlier
// appearance). Counts each language's net change once.
func technologyTransitions(timeline map[int][]string, years []int) int {
	sortInts(years)
	seenBefore := make(map[string]bool)
	everSeen := make(map[string]bool)
	transitions := 0
	for _, y := range years {
		langs := timeline[y]
		current := make(map[string]bool)
		for _, l := range langs {
			current[l] = true
			if !seenBefore[l] {
				transitions++ // adoption in this year
			}
			everSeen[l] = true
		}
		// Drop: a language seen before this year but absent now and never seen
		// again in a later year is counted when we reach the last year it appears.
		seenBefore = current
	}
	// Count drops: languages that appeared but are absent from the final year
	// (and thus "dropped" by the latest footprint).
	finalLangs := make(map[string]bool)
	for _, l := range timeline[years[len(years)-1]] {
		finalLangs[l] = true
	}
	for l := range everSeen {
		if !finalLangs[l] {
			transitions++
		}
	}
	return transitions
}

func technologyLevel(diversification int, specialization float64) Level {
	switch {
	case diversification <= 1:
		return LevelLow // single-language footprint
	case specialization >= 0.8:
		return LevelHigh // persistent focus despite breadth
	case diversification >= 5:
		return LevelBroad
	default:
		return LevelModerate
	}
}

func joinLang(langs []string) string {
	out := ""
	for i, l := range langs {
		if i > 0 {
			out += ", "
		}
		out += l
	}
	if out == "" {
		out = "none"
	}
	return out
}
