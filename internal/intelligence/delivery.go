package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// DeliveryIntelligence quantifies release discipline across the portfolio.
// Aggregated from per-repository Repository Intelligence.
type DeliveryIntelligence struct {
	dimensionBase
	Level            Level
	DisciplinedRatio float64
	ReleasedRepos    int
	TotalReleases    int
	Confidence       Confidence
}

func buildDelivery(repos []repositoryintelligence.RepositoryIntelligence) DeliveryIntelligence {
	dim := DeliveryIntelligence{}
	dim.name = "delivery"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Delivery is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("delivery", factItem("repositories", 0)),
		}
		return dim
	}

	var released, disciplined, totalReleases int
	for _, r := range repos {
		if r.Delivery.ReleaseCount > 0 {
			released++
			totalReleases += r.Delivery.ReleaseCount
		}
		if r.Delivery.Level == repositoryintelligence.LevelDisciplined {
			disciplined++
		}
	}

	ratio := float64(disciplined) / float64(total)
	dim.DisciplinedRatio = ratio
	dim.ReleasedRepos = released
	dim.TotalReleases = totalReleases
	dim.Level = levelFromRatio(ratio, 0.5, 0.2)
	dim.Confidence = confidenceForSample(len(repos))
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "delivery", []string{"delivery"},
			factItem("released repositories", released),
			factItem("repositories", total),
			factItem("disciplined repositories", disciplined),
			factItem("total releases", totalReleases),
		),
	}
	dim.summary = fmt.Sprintf("Delivery is %s (%d of %d repositories ship releases; %d total).", dim.Level, released, total, totalReleases)
	return dim
}
