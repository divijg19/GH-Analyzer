package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

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
	fmt.Println("Top candidates")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No matches found. Try lowering constraints.")
		return
	}

	printDetailedResultList(results, 1, false)
}

func printGroupedCandidates(results []engine.Result) {
	fmt.Println("Top candidates")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No candidates found. Try broader search input.")
		return
	}

	high := make([]engine.Result, 0, len(results))
	moderate := make([]engine.Result, 0, len(results))
	low := make([]engine.Result, 0, len(results))

	for _, result := range results {
		switch confidenceLabel(result.Score) {
		case "high":
			high = append(high, result)
		case "moderate":
			moderate = append(moderate, result)
		default:
			low = append(low, result)
		}
	}

	nextRank := 1
	if len(high) > 0 {
		fmt.Println("High confidence")
		nextRank = printDetailedResultList(high, nextRank, true)
	}
	if len(moderate) > 0 {
		fmt.Println("Moderate confidence")
		nextRank = printDetailedResultList(moderate, nextRank, true)
	}
	if len(low) > 0 {
		fmt.Println("Low confidence")
		_ = printDetailedResultList(low, nextRank, true)
	}
}

func printCompactResults(results []engine.Result) {
	if len(results) == 0 {
		fmt.Println("No results")
		return
	}

	nameWidth := maxUsernameWidth(results)
	for i, result := range results {
		fmt.Printf("%d. %-*s  %.2f\n", i+1, nameWidth, result.Profile.Username, result.Score)
	}
}

func printDetailedResultList(results []engine.Result, startRank int, includeConfidence bool) int {
	nameWidth := maxUsernameWidth(results)

	for i, result := range results {
		rank := startRank + i
		if includeConfidence {
			fmt.Printf("%d. %-*s  %.2f (%s)\n", rank, nameWidth, result.Profile.Username, result.Score, confidenceLabel(result.Score))
		} else {
			fmt.Printf("%d. %-*s  %.2f\n", rank, nameWidth, result.Profile.Username, result.Score)
		}
		fmt.Println("   Why:")
		for _, reason := range result.Reasons {
			fmt.Printf("   - %s\n", reason)
		}
		fmt.Println()
	}

	return startRank + len(results)
}

func maxUsernameWidth(results []engine.Result) int {
	width := 1
	for _, result := range results {
		if len(result.Profile.Username) > width {
			width = len(result.Profile.Username)
		}
	}

	return width
}

func printDatasetPreview(path string, indexData indexpkg.Index) {
	profiles := indexData.All()
	limit := 5
	if len(profiles) < limit {
		limit = len(profiles)
	}

	fmt.Printf("Dataset: %s\n", path)
	fmt.Printf("Profiles: %d\n", len(profiles))
	fmt.Println("Preview")

	if limit == 0 {
		fmt.Println("No profiles")
		return
	}

	preview := profiles[:limit]
	nameWidth := 1
	for _, profile := range preview {
		if len(profile.Username) > nameWidth {
			nameWidth = len(profile.Username)
		}
	}

	for i, profile := range preview {
		fmt.Printf("%d. %-*s  c:%.2f  o:%.2f  d:%.2f  a:%.2f\n",
			i+1,
			nameWidth,
			profile.Username,
			profile.Signals["consistency"],
			profile.Signals["ownership"],
			profile.Signals["depth"],
			profile.Signals["activity"],
		)
	}
}

func printDatasetInfo(path string, indexData indexpkg.Index) {
	fmt.Printf("Dataset: %s\n", path)
	fmt.Printf("Profiles: %d\n", len(indexData.All()))
}

func printProfileSignals(profile indexpkg.Profile) {
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
}

func printAverageSignals(indexData indexpkg.Index) {
	stats := calculateSignalStats(indexData)
	fmt.Printf("avg consistency: %.2f\n", stats.Avg["consistency"])
	fmt.Printf("avg ownership: %.2f\n", stats.Avg["ownership"])
}

func printDatasetSummary(path string, indexData indexpkg.Index) {
	stats := calculateSignalStats(indexData)

	fmt.Println("Dataset:")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Profiles: %d\n", len(indexData.All()))
	fmt.Printf("avg consistency: %.2f\n", stats.Avg["consistency"])
	fmt.Printf("avg ownership: %.2f\n", stats.Avg["ownership"])
	fmt.Printf("avg depth: %.2f\n", stats.Avg["depth"])
	fmt.Printf("avg activity: %.2f\n", stats.Avg["activity"])
}

func printDatasetStats(path string, indexData indexpkg.Index) {
	stats := calculateSignalStats(indexData)

	fmt.Printf("Dataset: %s\n", path)
	fmt.Printf("Profiles: %d\n", len(indexData.All()))
	fmt.Println()
	printSignalStatBlock("consistency", stats)
	printSignalStatBlock("ownership", stats)
	printSignalStatBlock("depth", stats)
	printSignalStatBlock("activity", stats)
}

type signalStats struct {
	Avg map[string]float64
	Min map[string]float64
	Max map[string]float64
}

func calculateSignalStats(indexData indexpkg.Index) signalStats {
	signals := []string{"consistency", "ownership", "depth", "activity"}
	profiles := indexData.All()

	stats := signalStats{
		Avg: map[string]float64{},
		Min: map[string]float64{},
		Max: map[string]float64{},
	}

	if len(profiles) == 0 {
		for _, signal := range signals {
			stats.Avg[signal] = 0
			stats.Min[signal] = 0
			stats.Max[signal] = 0
		}

		return stats
	}

	sums := map[string]float64{}
	for _, signal := range signals {
		sums[signal] = 0
		stats.Min[signal] = math.MaxFloat64
		stats.Max[signal] = -math.MaxFloat64
	}

	for _, profile := range profiles {
		for _, signal := range signals {
			value := profile.Signals[signal]
			sums[signal] += value
			if value < stats.Min[signal] {
				stats.Min[signal] = value
			}
			if value > stats.Max[signal] {
				stats.Max[signal] = value
			}
		}
	}

	count := float64(len(profiles))
	for _, signal := range signals {
		stats.Avg[signal] = sums[signal] / count
	}

	return stats
}

func printSignalStatBlock(signal string, stats signalStats) {
	fmt.Printf("%s:\n", signal)
	fmt.Printf("  avg: %.2f\n", stats.Avg[signal])
	fmt.Printf("  min: %.2f\n", stats.Min[signal])
	fmt.Printf("  max: %.2f\n", stats.Max[signal])
}

func confidenceLabel(score float64) string {
	switch {
	case score > 0.75:
		return "high"
	case score > 0.50:
		return "moderate"
	default:
		return "low"
	}
}
