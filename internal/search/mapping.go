package search

import (
	"fmt"
	"strings"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/presets"
	"github.com/divijg19/Atlas/internal/signals"
)

func MapIntent(input string) (engine.Query, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if normalized == "" {
		return engine.Query{}, fmt.Errorf("search input is required")
	}

	keywordConditions, expressionConditions, err := parseInput(normalized)
	if err != nil {
		return engine.Query{}, err
	}

	expressionSignals := map[string]struct{}{}
	for _, condition := range expressionConditions {
		expressionSignals[condition.Signal] = struct{}{}
	}

	conditions := make([]engine.Condition, 0, len(keywordConditions)+len(expressionConditions))
	for _, condition := range keywordConditions {
		if _, exists := expressionSignals[condition.Signal]; exists {
			continue
		}
		conditions = append(conditions, condition)
	}
	conditions = append(conditions, expressionConditions...)

	return engine.Query{Conditions: conditions}, nil
}

func parseInput(input string) ([]engine.Condition, []engine.Condition, error) {
	tokens := strings.Fields(input)
	keywordConditions := make([]engine.Condition, 0, len(tokens))
	expressionConditions := make([]engine.Condition, 0, len(tokens))
	seenKeywordSignals := map[string]struct{}{}

	for i := 0; i < len(tokens); {
		token := strings.TrimSpace(tokens[i])
		if token == "" {
			i++
			continue
		}

		if strings.EqualFold(token, "and") {
			i++
			continue
		}

		if mapped, ok := keywordConditionsForToken(token); ok {
			for _, condition := range mapped {
				if _, exists := seenKeywordSignals[condition.Signal]; exists {
					continue
				}
				keywordConditions = append(keywordConditions, condition)
				seenKeywordSignals[condition.Signal] = struct{}{}
			}
			i++
			continue
		}

		condition, span, matched, err := parseConditionFromTokens(tokens, i)
		if err != nil {
			return nil, nil, err
		}
		if matched {
			expressionConditions = append(expressionConditions, condition)
			i += span
			continue
		}

		i++
	}

	return keywordConditions, expressionConditions, nil
}

func keywordConditionsForToken(token string) ([]engine.Condition, bool) {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "backend":
		return []engine.Condition{{Signal: signals.SignalDepth, Operator: ">=", Value: 0.6}}, true
	case "frontend":
		return []engine.Condition{{Signal: signals.SignalDepth, Operator: ">=", Value: 0.5}}, true
	case "systems":
		return []engine.Condition{{Signal: signals.SignalDepth, Operator: ">=", Value: 0.8}}, true
	case "consistent":
		return []engine.Condition{{Signal: signals.SignalConsistency, Operator: ">=", Value: 0.7}}, true
	case "reliable":
		return []engine.Condition{{Signal: signals.SignalConsistency, Operator: ">=", Value: 0.8}, {Signal: signals.SignalActivity, Operator: ">=", Value: 1.0}}, true
	case "active":
		return []engine.Condition{{Signal: signals.SignalActivity, Operator: ">=", Value: 1.0}}, true
	case "beginner":
		return []engine.Condition{{Signal: signals.SignalDepth, Operator: "<=", Value: 0.4}}, true
	case "strong":
		query, err := presets.Preset("strong")
		if err != nil {
			return nil, false
		}

		return query.Conditions, true
	default:
		return nil, false
	}
}

func parseConditionFromTokens(tokens []string, start int) (engine.Condition, int, bool, error) {
	spans := []int{3, 2, 1}

	for _, span := range spans {
		end := start + span
		if end > len(tokens) {
			continue
		}

		candidate := strings.TrimSpace(strings.Join(tokens[start:end], " "))
		if candidate == "" {
			continue
		}

		condition, err := ParseCondition(candidate)
		if err == nil {
			return condition, span, true, nil
		}

		if LooksLikeCondition(candidate) {
			return engine.Condition{}, 0, false, fmt.Errorf("invalid condition %q", candidate)
		}
	}

	return engine.Condition{}, 0, false, nil
}
