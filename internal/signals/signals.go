package signals

func SignalsFromReport(report Report) map[string]float64 {
	consistency := clamp01(float64(report.Scores.Consistency) / scoreScale)
	ownership := clamp01(float64(report.Scores.Ownership) / scoreScale)
	depth := clamp01(float64(report.Scores.Depth) / scoreScale)
	activity := 0.0
	if consistency > 0 {
		activity = 1.0
	}

	signals := map[string]float64{
		"consistency": consistency,
		"ownership":   ownership,
		"depth":       depth,
		"activity":    activity,
	}

	return signals
}

func clamp01(value float64) float64 {
	if value < minSignalValue {
		return minSignalValue
	}
	if value > maxSignalValue {
		return maxSignalValue
	}

	return value
}
