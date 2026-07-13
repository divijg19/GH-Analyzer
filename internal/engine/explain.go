package engine

import (
	"fmt"
	"strings"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/indicators"
)

func explain(p index.Profile, q Query) []string {
	if len(q.Conditions) == 0 {
		return []string{"Ranked by overall signal strength"}
	}

	if p.Facts != nil {
		return explainWithEvidence(p, q)
	}

	return explainFromConditions(p, q)
}

func explainWithEvidence(p index.Profile, q Query) []string {
	sig := indicators.FromMap(p.Signals)

	allEvidence := evidence.GenerateEvidence(*p.Facts, sig)

	matched := map[string]struct{}{}
	for _, c := range q.Conditions {
		signal := strings.ToLower(strings.TrimSpace(c.Signal))
		if _, exists := matched[signal]; exists {
			continue
		}
		if match(p, c) {
			matched[signal] = struct{}{}
		}
	}

	var reasons []string
	for _, group := range allEvidence {
		if _, ok := matched[group.Signal]; !ok {
			continue
		}
		for _, item := range group.Items {
			reasons = append(reasons, item.Description)
		}
	}

	if len(reasons) == 0 {
		return []string{"Ranked by overall signal strength"}
	}

	return reasons
}

func explainFromConditions(p index.Profile, q Query) []string {
	reasons := make([]string, 0, len(q.Conditions))
	seenSignals := map[string]struct{}{}

	for _, condition := range q.Conditions {
		if !match(p, condition) {
			continue
		}

		signal := strings.ToLower(strings.TrimSpace(condition.Signal))
		reason := reasonForSignal(signal)
		if reason == "" {
			continue
		}
		if _, exists := seenSignals[signal]; exists {
			continue
		}

		reason = fmt.Sprintf("%s (%.2f %s %.2f)", reason, p.Signals[signal], strings.TrimSpace(condition.Operator), condition.Value)

		reasons = append(reasons, reason)
		seenSignals[signal] = struct{}{}
	}

	return reasons
}

func reasonForSignal(signal string) string {
	switch strings.ToLower(strings.TrimSpace(signal)) {
	case indicators.SignalConsistency:
		return "High consistency"
	case indicators.SignalOwnership:
		return "Strong ownership"
	case indicators.SignalDepth:
		return "Deep project work"
	case indicators.SignalActivity:
		return "Recently active"
	default:
		return ""
	}
}
