package observations

import "time"

// RepositoryVestige is the canonical observation of a repository. Every
// acquisition backend produces vestiges; Atlas never acquires from them
// directly. Fields are grouped into repository observation domains.
type RepositoryVestige struct {
	// Repository identity
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	Archived   bool   `json:"archived"`
	Template   bool   `json:"template"`

	// Repository ownership
	Fork              bool   `json:"fork"`
	ParentRepository  string `json:"parent_repository,omitempty"`
	CollaboratorCount int    `json:"collaborator_count"`

	// Repository timeline
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	LatestReleaseAt time.Time `json:"latest_release_at,omitempty"`
	ReleaseCount    int       `json:"release_count"`

	// Repository technology
	License                string           `json:"license"`
	Topics                 []string         `json:"topics"`
	DefaultBranch          string           `json:"default_branch"`
	DefaultBranchProtected bool             `json:"default_branch_protected"`
	LanguageDistribution   map[string]int64 `json:"language_distribution,omitempty"`

	// Repository maintenance
	OpenIssues       int `json:"open_issues_count"`
	Stars            int `json:"stargazers_count"`
	Forks            int `json:"forks_count"`
	Watchers         int `json:"watchers_count"`
	PullRequestCount int `json:"pull_request_count"`

	// Repository structure
	Size              int  `json:"size"`
	DiscussionEnabled bool `json:"discussion_enabled"`
}

// ObservationID returns the deterministic, Atlas-owned identity of this
// repository observation. Identity is constructed during normalization from the
// normalized repository name, which is unique within a candidate's portfolio and
// stable across executions. GitHub provides acquisition; Atlas owns identity.
func (v RepositoryVestige) ObservationID() string {
	return v.Name
}
