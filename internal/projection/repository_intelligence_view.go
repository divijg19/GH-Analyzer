package projection

import (
	"context"
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/observations"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// RepositoryDimensionView is a presentation-oriented view of a single
// repository intelligence dimension. Projection never recomputes meaning; the
// canonical data lives in repositoryintelligence.RepositoryIntelligence.
type RepositoryDimensionView struct {
	Name     string                   `json:"name"`
	Summary  string                   `json:"summary"`
	Evidence []evidence.EvidenceGroup `json:"evidence,omitempty"`
}

// RepositoryIntelligenceView is a presentation view of the deterministic
// interpretation of a single repository, across all thirteen dimensions.
type RepositoryIntelligenceView struct {
	Repository string                    `json:"repository"`
	Dimensions []RepositoryDimensionView `json:"dimensions"`
}

// RepositoryIntelligenceViews builds one view per repository from the raw
// vestiges, using the canonical repositoryintelligence layer. It is
// deterministic for a given referenceTime and never reinterprets values.
func RepositoryIntelligenceViews(repos []observations.RepositoryVestige, referenceTime time.Time) []RepositoryIntelligenceView {
	if len(repos) == 0 {
		return nil
	}
	views := make([]RepositoryIntelligenceView, 0, len(repos))
	for _, r := range repos {
		ri := repositoryintelligence.BuildRepositoryIntelligence(context.Background(), r, referenceTime)
		dims := ri.Dimensions()
		dimViews := make([]RepositoryDimensionView, 0, len(dims))
		for _, d := range dims {
			dimViews = append(dimViews, RepositoryDimensionView{
				Name:     d.Name(),
				Summary:  d.Summary(),
				Evidence: d.Evidence(),
			})
		}
		views = append(views, RepositoryIntelligenceView{
			Repository: ri.Repository,
			Dimensions: dimViews,
		})
	}
	return views
}
