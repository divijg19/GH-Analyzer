package acquisition

import obs "github.com/divijg19/Atlas/internal/observations"

// mergeVestiges combines a REST-derived RepositoryVestige with
// GraphQL-authoritative observations to produce the final vestige.
//
// Merge policy is derived from OBSERVATION_SPECIFICATION.md §4 (Observation
// Ownership Model). Each field's owner and canonical acquisition mechanism
// are documented there. REST-authoritative fields are never modified.
//
// GraphQL-authoritative fields (defined in OBSERVATION_SPECIFICATION.md):
//
//	Field                      Owner   Acquisition
//	LanguageDistribution       Atlas   GraphQL
//	ReleaseCount               Atlas   GraphQL
//	LatestReleaseAt            Atlas   GraphQL
//	PullRequestCount           Atlas   GraphQL
//	DiscussionEnabled          Atlas   GraphQL
//	ParentRepository           Atlas   GraphQL
//	CollaboratorCount          Atlas   GraphQL
//	DefaultBranchProtected     Atlas   GraphQL
//
// Merge is unconditionally additive for GraphQL-authoritative fields because
// they have no REST overlap. No precedence logic is needed. If a
// GraphQL-authoritative field is zero-valued (unobserved), it is explicitly
// set to its Go zero value — representing "not yet observed" per the
// observation specification.
func mergeVestiges(base, enrichment obs.RepositoryVestige) obs.RepositoryVestige {
	// GraphQL-authoritative — LanguageDistribution (§4, Tier 1)
	base.LanguageDistribution = enrichment.LanguageDistribution

	// GraphQL-authoritative — ReleaseCount, LatestReleaseAt (§4, Tier 1)
	base.ReleaseCount = enrichment.ReleaseCount
	base.LatestReleaseAt = enrichment.LatestReleaseAt

	// GraphQL-authoritative — PullRequestCount (§4, Tier 1)
	base.PullRequestCount = enrichment.PullRequestCount

	// GraphQL-authoritative — DiscussionEnabled (§4, Tier 1)
	base.DiscussionEnabled = enrichment.DiscussionEnabled

	// GraphQL-authoritative — ParentRepository (§4, Tier 1)
	base.ParentRepository = enrichment.ParentRepository

	// GraphQL-authoritative — CollaboratorCount (§4, Tier 1)
	base.CollaboratorCount = enrichment.CollaboratorCount

	// GraphQL-authoritative — DefaultBranchProtected (§4, Tier 1)
	base.DefaultBranchProtected = enrichment.DefaultBranchProtected

	return base
}
