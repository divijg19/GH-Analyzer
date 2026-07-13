package acquisition

import (
	"time"

	"github.com/divijg19/Atlas/internal/contributions"
	obs "github.com/divijg19/Atlas/internal/observations"
	"github.com/divijg19/Atlas/internal/profile"
)

// normalizeRepo maps a GitHub repository DTO to the domain repository model.
func normalizeRepo(dto RepoDTO) obs.RepositoryVestige {
	license := ""
	if dto.License != nil {
		license = dto.License.Key
	}

	return obs.RepositoryVestige{
		Name:          dto.Name,
		Fork:          dto.Fork,
		Size:          dto.Size,
		UpdatedAt:     parseTime(dto.UpdatedAt),
		Visibility:    dto.Visibility,
		Archived:      dto.Archived,
		Template:      dto.IsTemplate,
		License:       license,
		Topics:        dto.Topics,
		Stars:         dto.Stars,
		Forks:         dto.Forks,
		Watchers:      dto.Watchers,
		OpenIssues:    dto.OpenIssues,
		CreatedAt:     parseTime(dto.CreatedAt),
		PushedAt:      parseTime(dto.PushedAt),
		DefaultBranch: dto.DefaultBranch,
	}
}

// NormalizeRepos maps a slice of repository DTOs to domain repositories.
func NormalizeRepos(dtos []RepoDTO) []obs.RepositoryVestige {
	repos := make([]obs.RepositoryVestige, len(dtos))
	for i, dto := range dtos {
		repos[i] = normalizeRepo(dto)
	}
	return repos
}

// NormalizeUser maps a GitHub user DTO to the domain metadata model.
func NormalizeUser(dto *UserDTO) *profile.UserMetadata {
	if dto == nil {
		return nil
	}

	return &profile.UserMetadata{
		Name:      dto.Name,
		Bio:       dto.Bio,
		Location:  dto.Location,
		Company:   dto.Company,
		Followers: dto.Followers,
		Following: dto.Following,
		CreatedAt: parseTime(dto.CreatedAt),
	}
}

// NormalizeContributions maps the raw contribution DTO to the domain summary.
// Total contributions aggregate both pull requests and issues.
func NormalizeContributions(dto *ContributionsDTO) *contributions.Summary {
	if dto == nil {
		return nil
	}

	return &contributions.Summary{
		TotalContributions: dto.PullRequests + dto.Issues,
		TotalPullRequests:  dto.PullRequests,
		IssuesOpened:       dto.Issues,
	}
}

// normalizeGraphQLRepo maps a GraphQL repository DTO to GraphQL-authoritative
// RepositoryVestige fields. It returns the fields as a partial vestige for
// merge with the REST-derived vestige.
//
// Only fields documented as GraphQL-authoritative in
// OBSERVATION_SPECIFICATION.md are populated.
func normalizeGraphQLRepo(repo *graphQLRepo) obs.RepositoryVestige {
	if repo == nil {
		return obs.RepositoryVestige{}
	}

	var partial obs.RepositoryVestige

	if repo.Languages != nil {
		dist := make(map[string]int64, len(repo.Languages.Edges))
		for _, edge := range repo.Languages.Edges {
			if edge.Node.Name != "" {
				dist[edge.Node.Name] = edge.Size
			}
		}
		partial.LanguageDistribution = dist
	}

	if repo.Releases != nil {
		partial.ReleaseCount = repo.Releases.TotalCount
		if len(repo.Releases.Nodes) > 0 {
			partial.LatestReleaseAt = parseTime(repo.Releases.Nodes[0].CreatedAt)
		}
	}

	if repo.PullRequests != nil {
		partial.PullRequestCount = repo.PullRequests.TotalCount
	}

	if repo.HasDiscussionsEnabled != nil {
		partial.DiscussionEnabled = *repo.HasDiscussionsEnabled
	}

	if repo.Parent != nil && repo.Parent.Name != "" {
		partial.ParentRepository = repo.Parent.Owner.Login + "/" + repo.Parent.Name
	}

	if repo.Collaborators != nil {
		partial.CollaboratorCount = repo.Collaborators.TotalCount
	}

	if repo.BranchProtectionRules != nil {
		partial.DefaultBranchProtected = repo.BranchProtectionRules.TotalCount > 0
	}

	return partial
}

// parseTime parses an RFC3339 timestamp. An empty or unparseable value yields
// the zero time, mirroring prior decoding behavior for optional timestamps.
func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}

	return parsed
}
