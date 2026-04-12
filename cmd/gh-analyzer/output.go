package main

import (
	"encoding/json"
	"fmt"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
)

func writeJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func printFilters(query engine.Query, presetName string, limit int) {
	fmt.Println("Filters:")
	if presetName != "" {
		fmt.Printf("Preset: %s\n", presetName)
	}
	fmt.Printf("Limit: %d\n", limit)
	for _, condition := range query.Conditions {
		fmt.Printf("- %s %s %.2f\n", condition.Signal, condition.Operator, condition.Value)
	}
	fmt.Println()
}

func printMatchSummary(totalMatches, shownCount, limit int) {
	fmt.Printf("Matches: %d\n", totalMatches)
	if limit > 0 {
		fmt.Printf("Showing: %d of %d\n", shownCount, totalMatches)
	} else {
		fmt.Printf("Showing: %d\n", shownCount)
	}
	fmt.Println()
}

func printTopMatches(results []engine.Result) {
	fmt.Println("Top Matches:")
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
	averages := calculateAverageSignals(indexData)
	fmt.Printf("avg consistency: %.2f\n", averages["consistency"])
	fmt.Printf("avg ownership: %.2f\n", averages["ownership"])
}

func printDatasetSummary(path string, indexData indexpkg.Index) {
	averages := calculateAverageSignals(indexData)

	fmt.Println("Dataset:")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Profiles: %d\n", len(indexData.All()))
	fmt.Printf("avg consistency: %.2f\n", averages["consistency"])
	fmt.Printf("avg ownership: %.2f\n", averages["ownership"])
	fmt.Printf("avg depth: %.2f\n", averages["depth"])
	fmt.Printf("avg activity: %.2f\n", averages["activity"])
}

func calculateAverageSignals(indexData indexpkg.Index) map[string]float64 {
	profiles := indexData.All()
	if len(profiles) == 0 {
		return map[string]float64{
			"consistency": 0,
			"ownership":   0,
			"depth":       0,
			"activity":    0,
		}
	}

	sums := map[string]float64{
		"consistency": 0,
		"ownership":   0,
		"depth":       0,
		"activity":    0,
	}

	for _, profile := range profiles {
		sums["consistency"] += profile.Signals["consistency"]
		sums["ownership"] += profile.Signals["ownership"]
		sums["depth"] += profile.Signals["depth"]
		sums["activity"] += profile.Signals["activity"]
	}

	count := float64(len(profiles))
	for key, sum := range sums {
		sums[key] = sum / count
	}

	return sums
}
