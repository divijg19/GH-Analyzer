package evaluation

// Confidence represents the confidence level of a search score.
type Confidence string

const (
	High     Confidence = "high"
	Moderate Confidence = "moderate"
	Low      Confidence = "low"

	// Confidence thresholds classify a normalized score into a label.
	// Scores above highConfidenceThreshold are High; above
	// moderateConfidenceThreshold are Moderate; otherwise Low.
	highConfidenceThreshold     = 0.75
	moderateConfidenceThreshold = 0.50
)

// ClassifyConfidence maps a normalized score to a confidence label.
func ClassifyConfidence(score float64) Confidence {
	switch {
	case score > highConfidenceThreshold:
		return High
	case score > moderateConfidenceThreshold:
		return Moderate
	default:
		return Low
	}
}
