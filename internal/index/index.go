package index

import (
	"github.com/divijg19/Atlas/internal/contributions"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/observations"
	"github.com/divijg19/Atlas/internal/profile"
)

// Profile is the canonical candidate aggregate. It stores what Atlas knows
// about a candidate before evaluation: repository facts, indicators,
// activity facts, metadata, and contributions.
// It never stores scores, confidence, or evaluation results.

type Profile struct {
	Username      string                       `json:"username"`
	Signals       map[string]float64           `json:"signals"`
	Repositories  []observations.RepositoryVestige `json:"repositories,omitempty"`
	Facts         *facts.RepositoryFacts       `json:"facts,omitempty"`
	Metadata      *profile.UserMetadata        `json:"metadata,omitempty"`
	Contributions *contributions.Summary       `json:"contributions,omitempty"`
	ActivityFacts *facts.ActivityFacts         `json:"activity_facts,omitempty"`
}

type Index struct {
	Profiles []Profile `json:"profiles"`
}

func (i *Index) Add(p Profile) {
	i.Profiles = append(i.Profiles, p)
}

func (i *Index) All() []Profile {
	profiles := make([]Profile, len(i.Profiles))
	copy(profiles, i.Profiles)

	return profiles
}
