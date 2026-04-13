package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	"github.com/divijg19/GH-Analyzer/internal/presets"
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

func queryFromPreset(name string) (engine.Query, error) {
	preset, err := presets.Preset(strings.ToLower(strings.TrimSpace(name)))
	if err != nil {
		return engine.Query{}, err
	}

	conditions := make([]engine.Condition, 0, len(preset.Conditions))
	for _, condition := range preset.Conditions {
		conditions = append(conditions, engine.Condition{
			Signal:   condition.Signal,
			Operator: condition.Operator,
			Value:    condition.Value,
		})
	}

	return engine.Query{Conditions: conditions, Limit: preset.Limit}, nil
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
		return []engine.Condition{{Signal: "depth", Operator: ">=", Value: 0.6}}, true
	case "frontend":
		return []engine.Condition{{Signal: "depth", Operator: ">=", Value: 0.5}}, true
	case "systems":
		return []engine.Condition{{Signal: "depth", Operator: ">=", Value: 0.8}}, true
	case "consistent":
		return []engine.Condition{{Signal: "consistency", Operator: ">=", Value: 0.7}}, true
	case "reliable":
		return []engine.Condition{{Signal: "consistency", Operator: ">=", Value: 0.8}, {Signal: "activity", Operator: ">=", Value: 1.0}}, true
	case "active":
		return []engine.Condition{{Signal: "activity", Operator: ">=", Value: 1.0}}, true
	case "beginner":
		return []engine.Condition{{Signal: "depth", Operator: "<=", Value: 0.4}}, true
	case "strong":
		query, err := queryFromPreset("strong")
		if err != nil {
			return nil, false
		}

		conditions := make([]engine.Condition, 0, len(query.Conditions))
		conditions = append(conditions, query.Conditions...)
		return conditions, true
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

		condition, err := parseCondition(candidate)
		if err == nil {
			return condition, span, true, nil
		}

		if looksLikeCondition(candidate) {
			return engine.Condition{}, 0, false, fmt.Errorf("invalid condition %q", candidate)
		}
	}

	return engine.Condition{}, 0, false, nil
}

func parseCondition(raw string) (engine.Condition, error) {
	text := strings.TrimSpace(raw)
	operators := []string{">=", "<=", ">", "<"}

	for _, operator := range operators {
		idx := strings.Index(text, operator)
		if idx < 0 {
			continue
		}

		signal := strings.ToLower(strings.TrimSpace(text[:idx]))
		if signal == "" {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: missing signal before operator", raw)
		}
		if !isAllowedSignal(signal) {
			return engine.Condition{}, fmt.Errorf("invalid signal %q; expected consistency, ownership, depth, or activity", signal)
		}

		valueText := strings.TrimSpace(text[idx+len(operator):])
		if valueText == "" {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: missing value after operator", raw)
		}
		if startsWithOperator(valueText) {
			return engine.Condition{}, fmt.Errorf("invalid condition %q", raw)
		}

		value, err := strconv.ParseFloat(valueText, 64)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q", raw)
		}

		normalizedValue, err := normalizeThreshold(value)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: %w", raw, err)
		}

		return engine.Condition{Signal: signal, Operator: operator, Value: normalizedValue}, nil
	}

	return engine.Condition{}, fmt.Errorf("invalid condition %q", raw)
}

func normalizeThreshold(value float64) (float64, error) {
	if value > 1 {
		value = value / 100
	}
	if value < 0 || value > 1 {
		return 0, fmt.Errorf("must be between 0 and 1 (or 0 to 100)")
	}

	return value, nil
}

func isAllowedSignal(signal string) bool {
	switch signal {
	case "consistency", "ownership", "depth", "activity":
		return true
	default:
		return false
	}
}

func startsWithOperator(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	return strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "<") || strings.HasPrefix(trimmed, "=")
}

func looksLikeCondition(candidate string) bool {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return false
	}

	if strings.ContainsAny(trimmed, "<>=") {
		return true
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return false
	}

	return isAllowedSignal(fields[0])
}
