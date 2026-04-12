package search

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	"github.com/divijg19/GH-Analyzer/internal/presets"
)

var andSplitter = regexp.MustCompile(`(?i)\s+AND\s+`)

func MapIntent(input string) (engine.Query, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if normalized == "" {
		return engine.Query{}, fmt.Errorf("search input is required")
	}

	switch normalized {
	case "backend":
		return engine.Query{Conditions: []engine.Condition{{Signal: "depth", Operator: ">=", Value: 0.6}}}, nil
	case "consistent":
		return engine.Query{Conditions: []engine.Condition{{Signal: "consistency", Operator: ">=", Value: 0.7}}}, nil
	case "active":
		return engine.Query{Conditions: []engine.Condition{{Signal: "activity", Operator: ">=", Value: 1.0}}}, nil
	case "strong":
		return queryFromPreset("strong")
	default:
		conditions, err := parseExpression(normalized)
		if err != nil {
			return engine.Query{}, err
		}

		return engine.Query{Conditions: conditions}, nil
	}
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

func parseExpression(expression string) ([]engine.Condition, error) {
	parts := andSplitter.Split(expression, -1)
	conditions := make([]engine.Condition, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid condition expression: empty condition around AND")
		}

		condition, err := parseCondition(trimmed)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, condition)
	}

	return conditions, nil
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

		value, err := strconv.ParseFloat(valueText, 64)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: value must be a number", raw)
		}

		normalizedValue, err := normalizeThreshold(value)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: %w", raw, err)
		}

		return engine.Condition{Signal: signal, Operator: operator, Value: normalizedValue}, nil
	}

	return engine.Condition{}, fmt.Errorf("invalid condition %q; supported operators are >, >=, <, <=", raw)
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
