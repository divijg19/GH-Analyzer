package index

type Profile struct {
	Username string             `json:"username"`
	Signals  map[string]float64 `json:"signals"`
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
