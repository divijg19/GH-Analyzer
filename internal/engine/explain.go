package engine

import (
	"fmt"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

func Explain(p index.Profile, q Query) []string {
	if len(q.Conditions) == 0 {
		return []string{"Ranked by overall signal strength"}
	}

	reasons := make([]string, 0, len(q.Conditions))
	seen := map[string]struct{}{}

	for _, condition := range q.Conditions {
		if !Match(p, condition) {
			continue
		}

		signal := strings.ToLower(strings.TrimSpace(condition.Signal))
		reason := reasonForSignal(signal)
		if reason == "" {
			continue
		}

		reason = fmt.Sprintf("%s (%.2f)", reason, p.Signals[signal])

		if _, exists := seen[reason]; exists {
			continue
		}

		reasons = append(reasons, reason)
		seen[reason] = struct{}{}
	}

	return reasons
}

func reasonForSignal(signal string) string {
	switch strings.ToLower(strings.TrimSpace(signal)) {
	case "consistency":
		return "High consistency"
	case "ownership":
		return "Strong ownership"
	case "depth":
		return "Deep project work"
	case "activity":
		return "Recently active"
	default:
		return ""
	}
}
