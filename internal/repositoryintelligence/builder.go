package repositoryintelligence

import (
	"context"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

// repoContext holds deterministic, time-relative precomputations shared by all
// thirteen dimensions so that every dimension reasons over identical inputs.
type repoContext struct {
	vestige observations.RepositoryVestige
	ref     time.Time

	ageDays        int
	pushAgeDays    int
	updateAgeDays  int
	releaseAgeDays int

	languageCount int
	rankedLangs   []string
	primaryLang   string
	topicCount    int
	hasLicense    bool
}

func newRepoContext(v observations.RepositoryVestige, ref time.Time) repoContext {
	rc := repoContext{
		vestige:        v,
		ref:            ref,
		ageDays:        daysSince(v.CreatedAt, ref),
		pushAgeDays:    daysSince(v.PushedAt, ref),
		updateAgeDays:  daysSince(v.UpdatedAt, ref),
		releaseAgeDays: daysSince(v.LatestReleaseAt, ref),
		rankedLangs:    rankLanguages(v.LanguageDistribution),
		topicCount:     len(v.Topics),
		hasLicense:     v.License != "",
	}
	rc.languageCount = len(rc.rankedLangs)
	if rc.languageCount > 0 {
		rc.primaryLang = rc.rankedLangs[0]
	}
	return rc
}

// BuildRepositoryIntelligence produces the deterministic semantic interpretation
// of a single repository. It is pure: identical vestige and referenceTime yield a
// byte-identical RepositoryIntelligence.
func BuildRepositoryIntelligence(ctx context.Context, repo observations.RepositoryVestige, referenceTime time.Time) *RepositoryIntelligence {
	rc := newRepoContext(repo, referenceTime)
	return &RepositoryIntelligence{
		Repository:    repo.Name,
		ReferenceTime: referenceTime,
		Identity:      buildIdentity(rc),
		Health:        buildHealth(rc),
		Maintenance:   buildMaintenance(rc),
		Delivery:      buildDelivery(rc),
		Architecture:  buildArchitecture(rc),
		Technology:    buildTechnology(rc),
		Community:     buildCommunity(rc),
		Documentation: buildDocumentation(rc),
		Quality:       buildQuality(rc),
		Complexity:    buildComplexity(rc),
		Lifecycle:     buildLifecycle(rc),
		Governance:    buildGovernance(rc),
		Risk:          buildRisk(rc),
	}
}
