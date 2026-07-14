package repositoryintelligence

import (
	"fmt"
	"sort"
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/provenance"
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

func factItem(description string, value interface{}) evidence.Evidence {
	return evidence.Evidence{Kind: "fact", Description: description, Value: fmt.Sprintf("%v", value)}
}

// groupFrom builds an evidence group whose provenance references the given
// repository observation fields. Repository Intelligence consumes observations
// directly, so its provenance is observation-level: it names the exact vestige
// fields each dimension interpreted.
func groupFrom(rc repoContext, signal string, fields []string, items ...evidence.Evidence) evidence.EvidenceGroup {
	return evidence.EvidenceGroup{
		Signal: signal,
		Items:  items,
		Provenance: provenance.Chain{
			Observations: provenance.RepositoryObservations(rc.vestige.ObservationID(), fields...),
		},
	}
}

func daysSince(t time.Time, reference time.Time) int {
	if t.IsZero() {
		return 0
	}
	d := int(reference.Sub(t).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func rankLanguages(dist map[string]int64) []string {
	if len(dist) == 0 {
		return nil
	}
	type lp struct {
		name  string
		bytes int64
	}
	pairs := make([]lp, 0, len(dist))
	for l, b := range dist {
		pairs = append(pairs, lp{name: l, bytes: b})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].bytes != pairs[j].bytes {
			return pairs[i].bytes > pairs[j].bytes
		}
		return pairs[i].name < pairs[j].name
	})
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.name
	}
	return out
}
