package engine

import (
	"sort"

	"github.com/divijg19/Atlas/internal/evaluation"
	idx "github.com/divijg19/Atlas/internal/index"
)

type Result struct {
	Profile idx.Profile
	Score   float64
	Reasons []string
}

func Execute(index idx.Index, query Query, ranking rankingStrategy) []Result {
	if ranking == nil {
		ranking = evaluation.RankingPolicy{}
	}

	profiles := index.All()
	distribution := buildDistribution(profiles, ranking)
	results := make([]Result, 0, len(profiles))

	for _, profile := range profiles {
		if !matchesAllConditions(profile, query.Conditions) {
			continue
		}

		rawScore := ranking.Score(profile.Signals)

		results = append(results, Result{
			Profile: profile,
			Score:   calibrateScore(distribution, rawScore),
			Reasons: explain(profile, query),
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

func matchesAllConditions(profile idx.Profile, conditions []Condition) bool {
	for _, condition := range conditions {
		if !match(profile, condition) {
			return false
		}
	}

	return true
}
