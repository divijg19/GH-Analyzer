package intelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
)

func levelFromRatio(ratio, high, moderate float64) Level {
	switch {
	case ratio >= high:
		return LevelHigh
	case ratio >= moderate:
		return LevelModerate
	default:
		return LevelLow
	}
}

func pct(ratio float64) int {
	return int(ratio*100 + 0.5)
}

func factItem(description string, value interface{}) evidence.Evidence {
	return evidence.Evidence{Kind: "fact", Description: description, Value: fmt.Sprintf("%v", value)}
}

func group(signal string, items ...evidence.Evidence) evidence.EvidenceGroup {
	return evidence.EvidenceGroup{Signal: signal, Items: items}
}
