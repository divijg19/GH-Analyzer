package acquisition

const graphqlEndpoint = "https://api.github.com/graphql"

// repoQuery is the GraphQL query for enriching a single repository.
// It fetches all Tier 1 observation candidates that REST cannot supply.
const repoQuery = `query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    languages(first: 10, orderBy: {field: SIZE, direction: DESC}) {
      edges {
        size
        node { name }
      }
    }
    releases(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {
      totalCount
      nodes { createdAt }
    }
    pullRequests { totalCount }
    hasDiscussionsEnabled
    parent { name owner { login } }
    collaborators { totalCount }
    branchProtectionRules { totalCount }
  }
}`
