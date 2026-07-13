// Package contributions owns the contribution summary model consumed by
// Profile. Summary is a normalized observation type produced by acquisition.
//
// Contributions never acquires data, derives facts, or performs evaluation.
//
// Consumed by: index, projection.
package contributions

type Summary struct {
	TotalContributions int `json:"total_contributions"`
	TotalPullRequests  int `json:"total_pull_requests"`
	IssuesOpened       int `json:"issues_opened"`
}
