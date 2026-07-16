package facts

import "github.com/divijg19/Atlas/internal/provenance"

// repositoryFactObservations maps each repository fact to the repository
// observation fields it is deterministically derived from. Facts never duplicate
// observations; this table records which observed fields mattered. The mapping
// is static because the derivation in FromRepos is fixed and pure.
var repositoryFactObservations = map[string][]string{
	"total_repos":            {"Name"},
	"original_repos":         {"Fork"},
	"fork_repos":             {"Fork"},
	"recent_repos":           {"Fork", "UpdatedAt"},
	"deep_repos":             {"Fork", "Size"},
	"valid_repos":            {"Size"},
	"valid_original_repos":   {"Size", "Fork"},
	"largest_repo_size":      {"Size"},
	"latest_activity":        {"UpdatedAt"},
	"archived_repos":         {"Archived"},
	"template_repos":         {"Template"},
	"public_repos":           {"Visibility"},
	"private_repos":          {"Visibility"},
	"licensed_repos":         {"License"},
	"total_stars":            {"Stars"},
	"total_forks":            {"Forks"},
	"total_watchers":         {"Watchers"},
	"total_open_issues":      {"OpenIssues"},
	"total_topics":           {"Topics"},
	"language_count":         {"LanguageDistribution"},
	"ranked_languages":       {"LanguageDistribution"},
	"total_releases":         {"ReleaseCount"},
	"released_repos":         {"ReleaseCount"},
	"latest_release_at":      {"LatestReleaseAt"},
	"total_pull_requests":    {"PullRequestCount"},
	"total_collaborators":    {"CollaboratorCount"},
	"protected_branch_repos": {"DefaultBranchProtected"},
	"discussion_repos":       {"DiscussionEnabled"},
	"ranked_topics":          {"Topics"},
	"fork_lineage":           {"Fork", "ParentRepository"},

	// Activity facts derive from the activity observation family.
	"recent_commits":         {"Kind", "Count"},
	"recent_pull_requests":   {"Kind", "Count"},
	"recent_reviews":         {"Kind", "Count"},
	"recent_issues":          {"Kind", "Count"},
	"lifetime_commits":       {"Kind", "Count"},
	"lifetime_pull_requests": {"Kind", "Count"},
	"lifetime_reviews":       {"Kind", "Count"},
	"lifetime_issues":        {"Kind", "Count"},
	"contribution_breadth":   {"Kind"},
	"repository_breadth":     {"Repository"},
	"active_days":            {"ActiveDays"},
	"contribution_frequency": {"Kind", "Count", "ActiveDays"},
	"contribution_recency":   {"LifetimeTotal", "YearCount"},
	"activity_cadence":       {"ActiveDays", "TotalDays"},
	"repository_depth":       {"Repository", "Count"},
	"year_count":             {"Year"},
	"commit_cadence":         {"LifetimeCommits", "YearCount"},
	"technology_timeline":    {"CreatedAt", "LanguageDistribution"},
	"maintenance_buckets":    {"PushedAt"},
}

// FactProvenance returns the deterministic provenance chain for a single
// repository fact: the repository observation fields it was derived from. The
// returned chain references the repository observation family (no specific ID)
// because portfolio facts aggregate every repository observation.
func FactProvenance(name string) provenance.Chain {
	fields, ok := repositoryFactObservations[name]
	if !ok {
		return provenance.Chain{}
	}
	obs := make([]provenance.ObservationRef, 0, len(fields))
	for _, f := range fields {
		obs = append(obs, provenance.RepositoryFamilyField(f))
	}
	return provenance.Chain{
		Facts:        provenance.Facts(name),
		Observations: obs,
	}
}

// FactsProvenance merges the provenance of several facts into one chain,
// preserving argument order for determinism.
func FactsProvenance(names ...string) provenance.Chain {
	chains := make([]provenance.Chain, 0, len(names))
	for _, n := range names {
		chains = append(chains, FactProvenance(n))
	}
	return provenance.Merge(chains...)
}
