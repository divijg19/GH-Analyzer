package evaluation

// Confidence represents the confidence level of a search score.
type Confidence string

const (
	High     Confidence = "high"
	Moderate Confidence = "moderate"
	Low      Confidence = "low"
)

// ClassifyConfidence maps a normalized score to a confidence label.
func ClassifyConfidence(score float64) Confidence {
	switch {
	case score > 0.75:
		return High
	case score > 0.50:
		return Moderate
	default:
		return Low
	}
}
