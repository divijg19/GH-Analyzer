package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	"github.com/divijg19/GH-Analyzer/internal/presets"
)

var andSplitter = regexp.MustCompile(`(?i)\s+AND\s+`)

func runQuery(args []string) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printQueryHelp(fs.Output()) }
	consistency := fs.Float64("consistency", -1, "consistency threshold")
	ownership := fs.Float64("ownership", -1, "ownership threshold")
	depth := fs.Float64("depth", -1, "depth threshold")
	limit := fs.Int("limit", defaultQueryLimit, "max results")
	kLimit := fs.Int("k", defaultQueryLimit, "max results (alias)")
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	presetName := fs.String("preset", "", "preset name: strong, consistent, deep")
	jsonOutput := fs.Bool("json", false, "output JSON")
	compactOutput := fs.Bool("compact", false, "compact output (no explanations)")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	activePreset := strings.ToLower(strings.TrimSpace(*presetName))
	conditions := make([]engine.Condition, 0, 8)

	if activePreset != "" {
		presetConditions, err := conditionsFromPreset(activePreset)
		if err != nil {
			return fmt.Errorf("invalid preset: %w", err)
		}
		conditions = append(conditions, presetConditions...)
	}

	if *consistency >= 0 {
		value, err := normalizeThreshold(*consistency)
		if err != nil {
			return fmt.Errorf("invalid --consistency: %w", err)
		}
		conditions = append(conditions, engine.Condition{Signal: "consistency", Operator: ">=", Value: value})
	}
	if *ownership >= 0 {
		value, err := normalizeThreshold(*ownership)
		if err != nil {
			return fmt.Errorf("invalid --ownership: %w", err)
		}
		conditions = append(conditions, engine.Condition{Signal: "ownership", Operator: ">=", Value: value})
	}
	if *depth >= 0 {
		value, err := normalizeThreshold(*depth)
		if err != nil {
			return fmt.Errorf("invalid --depth: %w", err)
		}
		conditions = append(conditions, engine.Condition{Signal: "depth", Operator: ">=", Value: value})
	}

	expression := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if expression != "" {
		exprConditions, err := parseExpression(expression)
		if err != nil {
			return err
		}
		conditions = append(conditions, exprConditions...)
	}

	if len(conditions) == 0 {
		activePreset = defaultQueryPreset
		presetConditions, err := conditionsFromPreset(activePreset)
		if err != nil {
			return err
		}
		conditions = append(conditions, presetConditions...)
	}

	resolvedLimit, err := resolveLimitFlag(fs, *limit, *kLimit)
	if err != nil {
		return err
	}

	if resolvedLimit < 0 {
		return fmt.Errorf("invalid --limit: must be >= 0")
	}

	runner := engine.New(engine.WeightedRanking{})
	fullQuery := engine.Query{Conditions: conditions, Limit: 0}
	totalMatches := len(runner.Query(indexData, fullQuery))

	query := fullQuery
	query.Limit = resolvedLimit
	results := runner.Query(indexData, query)
	if *jsonOutput {
		return writeJSON(results)
	}

	printFilters(query, activePreset, resolvedLimit)
	printMatchSummary(totalMatches, len(results), resolvedLimit)
	if *compactOutput {
		printCompactResults(results)
	} else {
		printTopMatches(results)
	}

	return nil
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

func conditionsFromPreset(name string) ([]engine.Condition, error) {
	presetQuery, err := presets.Preset(name)
	if err != nil {
		return nil, err
	}

	conditions := make([]engine.Condition, 0, len(presetQuery.Conditions))
	for _, condition := range presetQuery.Conditions {
		conditions = append(conditions, engine.Condition{
			Signal:   condition.Signal,
			Operator: condition.Operator,
			Value:    condition.Value,
		})
	}

	return conditions, nil
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

		normalized, err := normalizeThreshold(value)
		if err != nil {
			return engine.Condition{}, fmt.Errorf("invalid condition %q: %w", raw, err)
		}

		return engine.Condition{Signal: signal, Operator: operator, Value: normalized}, nil
	}

	return engine.Condition{}, fmt.Errorf("invalid condition %q; supported operators are >, >=, <, <=", raw)
}

func isAllowedSignal(signal string) bool {
	switch signal {
	case "consistency", "ownership", "depth", "activity":
		return true
	default:
		return false
	}
}
