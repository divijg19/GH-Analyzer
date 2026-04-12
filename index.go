package ghanalyzer

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
