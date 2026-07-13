package engine

// rankingStrategy scores a candidate profile for search ranking. The concrete
// policy is owned by the evaluation layer; engine only defines the contract.
type rankingStrategy interface {
	Score(map[string]float64) float64
}
