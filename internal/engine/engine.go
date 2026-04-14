package engine

import (
	"sort"

	idx "github.com/divijg19/GH-Analyzer/internal/index"
)

type Result struct {
	Profile idx.Profile
	Score   float64
	Reasons []string
}

type Engine struct {
	ranking RankingStrategy
}

func New(ranking RankingStrategy) *Engine {
	if ranking == nil {
		ranking = WeightedRanking{}
	}

	return &Engine{ranking: ranking}
}

func (e *Engine) Query(index idx.Index, query Query) []Result {
	return Execute(index, query, e.ranking)
}

func Execute(index idx.Index, query Query, ranking RankingStrategy) []Result {
	if ranking == nil {
		ranking = WeightedRanking{}
	}

	profiles := index.All()
	distribution := BuildDistribution(profiles, ranking)
	results := make([]Result, 0, len(profiles))

	for _, profile := range profiles {
		if !matchesAllConditions(profile, query.Conditions) {
			continue
		}

		rawScore := ranking.Score(profile)

		results = append(results, Result{
			Profile: profile,
			Score:   CalibrateScore(distribution, rawScore),
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

func matchesAllConditions(profile idx.Profile, conditions []Condition) bool {
	for _, condition := range conditions {
		if !Match(profile, condition) {
			return false
		}
	}

	return true
}
