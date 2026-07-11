package index

import (
	"github.com/divijg19/Atlas/internal/contributions"
	"github.com/divijg19/Atlas/internal/profile"
	"github.com/divijg19/Atlas/internal/signals"
)

type Profile struct {
	Username      string                   `json:"username"`
	Signals       map[string]float64       `json:"signals"`
	Facts         *signals.RepositoryFacts `json:"facts,omitempty"`
	Metadata      *profile.UserMetadata    `json:"metadata,omitempty"`
	Contributions *contributions.Summary   `json:"contributions,omitempty"`
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
