package signals

// SignalsToMap converts a Signals struct to a map[string]float64
// suitable for storage in Profile and use by the query engine.
func SignalsToMap(s Signals) map[string]float64 {
	return map[string]float64{
		"ownership":   clamp01(s.Ownership),
		"consistency": clamp01(s.Consistency),
		"depth":       clamp01(s.Depth),
		"activity":    clamp01(s.Activity),
	}
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
