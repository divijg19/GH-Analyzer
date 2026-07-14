package intelligence

import (
	"context"
	"errors"
	"time"

	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// BuildCandidateIntelligence produces the deterministic semantic interpretation
// of a Profile. It is pure: identical Profile and referenceTime yield a
// byte-identical CandidateIntelligence.
//
// Candidate Intelligence is an aggregation layer: it derives every portfolio
// dimension by aggregating the per-repository RepositoryIntelligence of each
// repository in the Profile. It never reads RepositoryVestige facts directly for
// these dimensions; Repository Intelligence is the single owner of repository
// interpretation. See docs/REPOSITORY_INTELLIGENCE.md §7.
func BuildCandidateIntelligence(ctx context.Context, p *index.Profile, referenceTime time.Time) (*CandidateIntelligence, error) {
	if p == nil {
		return nil, errors.New("intelligence: nil profile")
	}

	repos := make([]repositoryintelligence.RepositoryIntelligence, 0, len(p.Repositories))
	for _, r := range p.Repositories {
		repos = append(repos, *repositoryintelligence.BuildRepositoryIntelligence(ctx, r, referenceTime))
	}

	ci := &CandidateIntelligence{
		Username:      p.Username,
		ReferenceTime: referenceTime,
		Ownership:     buildOwnership(repos),
		Delivery:      buildDelivery(repos),
		Breadth:       buildBreadth(repos),
		Focus:         buildFocus(repos),
		Maintenance:   buildMaintenance(repos),
		Project:       buildProject(repos),
		Collaboration: buildCollaboration(repos),
		Documentation: buildDocumentation(repos),
		Portfolio:     buildPortfolio(repos, p.Metadata, referenceTime),
	}
	return ci, nil
}
