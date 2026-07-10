package projection

import (
	"github.com/divijg19/GH-Analyzer/internal/evaluation"
	"github.com/divijg19/GH-Analyzer/internal/index"
)

// SearchProjection is the canonical search presentation model.
// It owns the presentation shape for search results.
type SearchProjection struct {
	Username   string                `json:"username"`
	Score      float64               `json:"score"`
	Confidence evaluation.Confidence `json:"confidence"`
	Signals    map[string]float64    `json:"signals"`
	Reasons    []string              `json:"reasons"`
}

// BuildSearchProjection creates a search view from presentation-ready data.
// Projection receives evaluated data, never raw engine output.
func BuildSearchProjection(
	profile index.Profile,
	score float64,
	confidence evaluation.Confidence,
	reasons []string,
) SearchProjection {
	return SearchProjection{
		Username:   profile.Username,
		Score:      score,
		Confidence: confidence,
		Signals:    cloneSignals(profile.Signals),
		Reasons:    append([]string{}, reasons...),
	}
}

// cloneSignals creates a shallow copy of a signals map.
func cloneSignals(in map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
