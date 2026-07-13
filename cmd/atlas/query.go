package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/evaluation"
	"github.com/divijg19/Atlas/internal/presets"
	"github.com/divijg19/Atlas/internal/projection"
	searchpkg "github.com/divijg19/Atlas/internal/search"
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
		presetQuery, err := presets.Preset(activePreset)
		if err != nil {
			return fmt.Errorf("invalid preset: %w", err)
		}
		conditions = append(conditions, presetQuery.Conditions...)
	}

	if *consistency >= 0 {
		value, err := searchpkg.NormalizeThreshold(*consistency)
		if err != nil {
			return fmt.Errorf("invalid --consistency: %w", err)
		}
		conditions = append(conditions, engine.Condition{Signal: "consistency", Operator: ">=", Value: value})
	}
	if *ownership >= 0 {
		value, err := searchpkg.NormalizeThreshold(*ownership)
		if err != nil {
			return fmt.Errorf("invalid --ownership: %w", err)
		}
		conditions = append(conditions, engine.Condition{Signal: "ownership", Operator: ">=", Value: value})
	}
	if *depth >= 0 {
		value, err := searchpkg.NormalizeThreshold(*depth)
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
		presetQuery, err := presets.Preset(activePreset)
		if err != nil {
			return err
		}
		conditions = append(conditions, presetQuery.Conditions...)
	}

	resolvedLimit, err := resolveLimitFlag(fs, *limit, *kLimit)
	if err != nil {
		return err
	}

	if resolvedLimit < 0 {
		return fmt.Errorf("invalid --limit: must be >= 0")
	}

	fullQuery := engine.Query{Conditions: conditions, Limit: 0}
	totalMatches := len(engine.Execute(indexData, fullQuery, evaluation.RankingPolicy{}))

	query := fullQuery
	query.Limit = resolvedLimit
	results := engine.Execute(indexData, query, evaluation.RankingPolicy{})

	projections := make([]projection.SearchProjection, len(results))
	for i, result := range results {
		projections[i] = projection.BuildSearchProjection(
			result.Profile,
			result.Score,
			evaluation.ClassifyConfidence(result.Score),
			result.Reasons,
		)
	}

	if *jsonOutput {
		return writeJSON(projections)
	}

	printFilters(query, activePreset, resolvedLimit)
	printMatchSummary(totalMatches, len(projections), resolvedLimit)
	if *compactOutput {
		printCompactSearchProjections(projections)
	} else {
		printTopSearchProjections(projections)
	}

	return nil
}

func parseExpression(expression string) ([]engine.Condition, error) {
	parts := andSplitter.Split(expression, -1)
	conditions := make([]engine.Condition, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid condition expression: empty condition around AND")
		}

		condition, err := searchpkg.ParseCondition(trimmed)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, condition)
	}

	return conditions, nil
}

func parseCondition(raw string) (engine.Condition, error) {
	return searchpkg.ParseCondition(raw)
}
