package ghanalyzer

import "sort"

type Result struct {
	Profile Profile
	Score   float64
	Reasons []string
}

func Execute(index Index, query Query, ranking RankingStrategy) []Result {
	if ranking == nil {
		ranking = WeightedRanking{}
	}

	profiles := index.All()
	results := make([]Result, 0, len(profiles))

	for _, profile := range profiles {
		if !matchesAllConditions(profile, query.Conditions) {
			continue
		}

		results = append(results, Result{
			Profile: profile,
			Score:   ranking.Score(profile),
			Reasons: Explain(profile, query),
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Profile.Username < results[j].Profile.Username
		}

		return results[i].Score > results[j].Score
	})

	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results
}

func matchesAllConditions(profile Profile, conditions []Condition) bool {
	for _, condition := range conditions {
		if !Match(profile, condition) {
			return false
		}
	}

	return true
}
