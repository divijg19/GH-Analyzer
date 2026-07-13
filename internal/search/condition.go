// Package search owns query parsing and intent mapping for the search path.
//
// It translates natural-language keywords and structured conditions into
// engine.Query values consumed by the engine layer.
//
// Search never acquires observations, derives facts, computes indicators,
// generates evidence, evaluates candidates, or performs presentation.
//
// Consumed by: cmd/atlas.
package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/indicators"
)

func ParseCondition(raw string) (engine.Condition, error) {
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

		normalizedValue, err := NormalizeThreshold(value)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: %w", raw, err)
		}

		return engine.Condition{Signal: signal, Operator: operator, Value: normalizedValue}, nil
	}

	return engine.Condition{}, fmt.Errorf("invalid condition %q; supported operators are >, >=, <, <=", raw)
}

func NormalizeThreshold(value float64) (float64, error) {
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
	case indicators.SignalConsistency, indicators.SignalOwnership, indicators.SignalDepth, indicators.SignalActivity:
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

func LooksLikeCondition(candidate string) bool {
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
