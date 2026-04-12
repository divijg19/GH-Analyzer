package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

const (
	minReposForFullScore         = 3
	smallSampleOverallMultiplier = 0.7
	defaultDatasetPath           = "dataset.json"
	defaultQueryLimit            = 10
	defaultQueryPreset           = "strong"
)

var andSplitter = regexp.MustCompile(`(?i)\s+AND\s+`)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.TrimSpace(os.Args[1])
	args := os.Args[2:]

	var err error
	switch command {
	case "build":
		err = runBuild(args)
	case "query":
		err = runQuery(args)
	case "inspect":
		err = runInspect(args)
	default:
		err = runAnalyze(command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runAnalyze(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("missing GitHub username")
	}

	repos, err := signals.FetchRepos(username)
	if err != nil {
		return err
	}

	signalValues := signals.ExtractSignals(repos)

	scores := signals.ScoreSignals(signalValues)
	if len(repos) < minReposForFullScore {
		scores.Overall = int(math.Round(float64(scores.Overall) * smallSampleOverallMultiplier))
	}
	report := signals.BuildReport(username, scores, repos)

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Report for %s\n\n", report.Username)
	fmt.Println(string(output))
	return nil
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	filePath := fs.String("file", "", "path to username list file")
	outPath := fs.String("out", "", "output dataset path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	usernames, err := collectUsernames(fs.Args(), *filePath)
	if err != nil {
		return err
	}

	indexData, err := indexpkg.Build(usernames)
	if err != nil {
		return err
	}

	fmt.Printf("Built index with %d profiles\n", len(indexData.All()))
	printAverageSignals(indexData)

	resolvedOutPath := strings.TrimSpace(*outPath)
	if resolvedOutPath == "" {
		resolvedOutPath = defaultDatasetPath
	}

	if err := storage.Save(resolvedOutPath, indexData); err != nil {
		return err
	}
	fmt.Printf("Saved to %s\n", resolvedOutPath)

	return nil
}

func runQuery(args []string) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	consistency := fs.Float64("consistency", -1, "consistency threshold")
	ownership := fs.Float64("ownership", -1, "ownership threshold")
	depth := fs.Float64("depth", -1, "depth threshold")
	limit := fs.Int("limit", defaultQueryLimit, "max results")
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	presetName := fs.String("preset", "", "preset name: strong, consistent, deep")

	if err := fs.Parse(args); err != nil {
		return err
	}

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}

	conditions := make([]engine.Condition, 0, 8)

	if strings.TrimSpace(*presetName) != "" {
		preset, err := engine.QueryFromPreset(strings.ToLower(strings.TrimSpace(*presetName)))
		if err != nil {
			return fmt.Errorf("invalid preset: %w", err)
		}
		conditions = append(conditions, preset.Conditions...)
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
		preset, err := engine.QueryFromPreset(defaultQueryPreset)
		if err != nil {
			return err
		}
		conditions = append(conditions, preset.Conditions...)
	}

	if *limit < 0 {
		return fmt.Errorf("invalid --limit: must be >= 0")
	}

	query := engine.Query{Conditions: conditions, Limit: *limit}
	runner := engine.New(engine.WeightedRanking{})
	results := runner.Query(indexData, query)

	printFilters(query)
	fmt.Printf("Matches: %d\n\n", len(results))
	printTopMatches(results)

	return nil
}

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(fs.Args()) < 1 {
		return fmt.Errorf("username is required")
	}

	username := strings.TrimSpace(fs.Args()[0])
	if username == "" {
		return fmt.Errorf("username is required")
	}

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}

	profile, found := findProfile(indexData, username)
	if !found {
		return fmt.Errorf("profile %q not found", username)
	}

	fmt.Printf("Profile: %s\n", profile.Username)
	fmt.Println("Signals:")

	names := make([]string, 0, len(profile.Signals))
	for name := range profile.Signals {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("- %s: %.2f\n", name, profile.Signals[name])
	}

	return nil
}

func collectUsernames(positional []string, filePath string) ([]string, error) {
	usernames := append([]string{}, positional...)

	if strings.TrimSpace(filePath) != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			name := strings.TrimSpace(scanner.Text())
			if name == "" {
				continue
			}
			usernames = append(usernames, name)
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	clean := make([]string, 0, len(usernames))
	for _, name := range usernames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}

	if len(clean) == 0 {
		return nil, fmt.Errorf("no usernames provided")
	}

	return clean, nil
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

func printFilters(query engine.Query) {
	fmt.Println("Filters:")
	for _, condition := range query.Conditions {
		fmt.Printf("- %s %s %.2f\n", condition.Signal, condition.Operator, condition.Value)
	}
	fmt.Println()
}

func printTopMatches(results []engine.Result) {
	fmt.Println("Top matches")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No matches found. Try lowering constraints.")
		return
	}

	for i, result := range results {
		fmt.Printf("%d. %s - %.2f\n", i+1, result.Profile.Username, result.Score)
		fmt.Println("   Why:")
		for _, reason := range result.Reasons {
			fmt.Printf("   - %s\n", reason)
		}
		fmt.Println()
	}
}

func printAverageSignals(indexData indexpkg.Index) {
	profiles := indexData.All()
	if len(profiles) == 0 {
		fmt.Println("avg consistency: 0.00")
		fmt.Println("avg ownership: 0.00")
		return
	}

	var sumConsistency float64
	var sumOwnership float64

	for _, profile := range profiles {
		sumConsistency += profile.Signals["consistency"]
		sumOwnership += profile.Signals["ownership"]
	}

	count := float64(len(profiles))
	fmt.Printf("avg consistency: %.2f\n", sumConsistency/count)
	fmt.Printf("avg ownership: %.2f\n", sumOwnership/count)
}

func findProfile(indexData indexpkg.Index, username string) (indexpkg.Profile, bool) {
	for _, profile := range indexData.All() {
		if strings.EqualFold(profile.Username, username) {
			return profile, true
		}
	}

	return indexpkg.Profile{}, false
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gh-analyzer <username>")
	fmt.Println("  gh-analyzer build user1 user2 [--file users.txt] [--out dataset.json]")
	fmt.Println("  gh-analyzer query [--consistency 0.7] [--ownership 0.6] [--depth 0.5] [--limit 10] [--dataset dataset.json] [--preset strong] [expression]")
	fmt.Println("  gh-analyzer inspect username [--dataset dataset.json]")
	fmt.Println("  build saves to dataset.json by default")
}

func missingDatasetError(path string) error {
	if path == defaultDatasetPath {
		return fmt.Errorf("dataset %q not found. Run: gh-analyzer build <usernames>", path)
	}

	return fmt.Errorf("dataset %q not found. Run: gh-analyzer build <usernames> --out %s", path, path)
}
