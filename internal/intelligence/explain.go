package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/provenance"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

func levelFromRatio(ratio, high, moderate float64) Level {
	switch {
	case ratio >= high:
		return LevelHigh
	case ratio >= moderate:
		return LevelModerate
	default:
		return LevelLow
	}
}

func pct(ratio float64) int {
	return int(ratio*100 + 0.5)
}

// confidenceForSample derives a dimension's confidence from the volume of
// evidence supporting it: an empty or tiny sample yields low confidence, a
// moderate sample yields moderate, and a substantial sample yields high. This
// makes Confidence answer "how much evidence supports this conclusion?" rather
// than being a constant. The thresholds are documented calibration constants,
// not magic numbers.
const (
	confidenceLowSample      = 2
	confidenceModerateSample = 10
)

func confidenceForSample(n int) Confidence {
	switch {
	case n <= confidenceLowSample:
		return ConfidenceLow
	case n < confidenceModerateSample:
		return ConfidenceModerate
	default:
		return ConfidenceHigh
	}
}

func factItem(description string, value interface{}) evidence.Evidence {
	return evidence.Evidence{Kind: "fact", Description: description, Value: fmt.Sprintf("%v", value)}
}

func group(signal string, items ...evidence.Evidence) evidence.EvidenceGroup {
	return evidence.EvidenceGroup{Signal: signal, Items: items}
}

// repoRefs records which repositories contributed to a portfolio conclusion and,
// for each, which repository intelligence dimensions were consumed. Aggregation
// reorganizes knowledge; it never flattens it. Following a RepositoryRef back to
// the repository intelligence view yields that repository's full provenance
// chain without duplicating it here.
func repoRefs(repos []repositoryintelligence.RepositoryIntelligence, dims ...string) []provenance.RepositoryRef {
	if len(repos) == 0 || len(dims) == 0 {
		return nil
	}
	refs := make([]provenance.RepositoryRef, 0, len(repos)*len(dims))
	for _, r := range repos {
		for _, d := range dims {
			refs = append(refs, provenance.RepositoryRef{Repository: r.Repository, Dimension: d})
		}
	}
	return refs
}

// groupFrom builds a candidate evidence group whose provenance names the
// contributing repositories and the repository dimensions consumed.
func groupFrom(repos []repositoryintelligence.RepositoryIntelligence, signal string, dims []string, items ...evidence.Evidence) evidence.EvidenceGroup {
	return evidence.EvidenceGroup{
		Signal:     signal,
		Items:      items,
		Provenance: provenance.Chain{Repositories: repoRefs(repos, dims...)},
	}
}
