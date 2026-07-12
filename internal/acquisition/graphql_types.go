package acquisition

// graphQLRequest is the GraphQL API request envelope.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// graphQLResponse is the GraphQL API response envelope.
type graphQLResponse struct {
	Data   graphQLData `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// graphQLData contains the repository query result.
type graphQLData struct {
	Repository *graphQLRepo `json:"repository"`
}

// graphQLRepo mirrors the repository object from the GitHub GraphQL API.
type graphQLRepo struct {
	Languages             *graphQLLanguages      `json:"languages"`
	Releases              *graphQLCountWithNodes `json:"releases"`
	PullRequests          *graphQLCount          `json:"pullRequests"`
	HasDiscussionsEnabled *bool                  `json:"hasDiscussionsEnabled"`
	Parent                *graphQLParent         `json:"parent"`
	Collaborators         *graphQLCount          `json:"collaborators"`
	BranchProtectionRules *graphQLCount          `json:"branchProtectionRules"`
}

type graphQLLanguages struct {
	Edges []graphQLLanguageEdge `json:"edges"`
}

type graphQLLanguageEdge struct {
	Size int64           `json:"size"`
	Node graphQLLanguage `json:"node"`
}

type graphQLLanguage struct {
	Name string `json:"name"`
}

type graphQLCount struct {
	TotalCount int `json:"totalCount"`
}

type graphQLCountWithNodes struct {
	TotalCount int               `json:"totalCount"`
	Nodes      []graphQLNodeTime `json:"nodes"`
}

type graphQLNodeTime struct {
	CreatedAt string `json:"createdAt"`
}

type graphQLParent struct {
	Name  string       `json:"name"`
	Owner graphQLOwner `json:"owner"`
}

type graphQLOwner struct {
	Login string `json:"login"`
}
