package live

import (
	"context"
)

// FetchLiveUsernames searches GitHub repositories and returns deduplicated
// owner usernames, up to acquisition.maxUsers.
func FetchLiveUsernames(ctx context.Context, query string) ([]string, error) {
	return Client.SearchRepositoryOwners(ctx, query)
}
