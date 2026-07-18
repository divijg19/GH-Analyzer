package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// CollaborationIntelligence quantifies external engagement and shared ownership
// across the portfolio. Aggregated from per-repository Repository Intelligence.
type CollaborationIntelligence struct {
	dimensionBase
	Level              Level
	TotalCollaborators int
	ReposWithCollab    int
	TotalForks         int
	TotalWatchers      int
	Confidence         Confidence
}

func buildCollaboration(repos []repositoryintelligence.RepositoryIntelligence) CollaborationIntelligence {
	dim := CollaborationIntelligence{}
	dim.name = "collaboration"

	total := len(repos)
	if total == 0 {
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
		dim.summary = "Collaboration is unknown (no repositories)."
		dim.evidence = []evidence.EvidenceGroup{
			group("collaboration", factItem("repositories", 0)),
		}
		return dim
	}

	var collaborators, withCollab, forks, watchers int
	for _, r := range repos {
		collaborators += r.Community.Collaborators
		if r.Community.Collaborators > 0 {
			withCollab++
		}
		forks += r.Community.Forks
		watchers += r.Community.Watchers
	}

	dim.TotalCollaborators = collaborators
	dim.ReposWithCollab = withCollab
	dim.TotalForks = forks
	dim.TotalWatchers = watchers

	switch {
	case collaborators >= 5 || watchers >= 50 || forks >= 10:
		dim.Level = LevelHigh
	case withCollab > 0 || forks > 0 || watchers > 0:
		dim.Level = LevelModerate
	default:
		dim.Level = LevelLow
	}
	dim.Confidence = confidenceForSample(len(repos))
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "collaboration", []string{"community"},
			factItem("total collaborators", collaborators),
			factItem("repositories with collaborators", withCollab),
			factItem("total forks", forks),
			factItem("total watchers", watchers),
		),
	}
	dim.summary = fmt.Sprintf("Collaboration is %s (%d collaborators across %d repositories).", dim.Level, collaborators, withCollab)
	return dim
}
