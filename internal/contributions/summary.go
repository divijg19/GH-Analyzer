package contributions

type Summary struct {
	TotalContributions int `json:"total_contributions"`
	TotalPullRequests  int `json:"total_pull_requests"`
	IssuesOpened       int `json:"issues_opened"`
}
