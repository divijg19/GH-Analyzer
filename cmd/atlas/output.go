package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/divijg19/Atlas/internal/engine"
	indexpkg "github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/projection"
	"github.com/divijg19/Atlas/internal/provenance"
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

func printTopSearchProjections(projections []projection.SearchProjection) {
	fmt.Println("Top candidates")
	fmt.Println()

	if len(projections) == 0 {
		fmt.Println("No matches found. Try lowering constraints.")
		return
	}

	printDetailedProjectionList(projections, 1, false)
}

func printGroupedSearchProjections(projections []projection.SearchProjection) {
	fmt.Println("Top candidates")
	fmt.Println()

	if len(projections) == 0 {
		fmt.Println("No candidates found. Try broader search input.")
		return
	}

	high := make([]projection.SearchProjection, 0, len(projections))
	moderate := make([]projection.SearchProjection, 0, len(projections))
	low := make([]projection.SearchProjection, 0, len(projections))

	for _, proj := range projections {
		switch proj.Confidence {
		case "high":
			high = append(high, proj)
		case "moderate":
			moderate = append(moderate, proj)
		default:
			low = append(low, proj)
		}
	}

	nextRank := 1
	if len(high) > 0 {
		fmt.Println("High confidence")
		nextRank = printDetailedProjectionList(high, nextRank, true)
	}
	if len(moderate) > 0 {
		fmt.Println("Moderate confidence")
		nextRank = printDetailedProjectionList(moderate, nextRank, true)
	}
	if len(low) > 0 {
		fmt.Println("Low confidence")
		_ = printDetailedProjectionList(low, nextRank, true)
	}
}

func printCompactSearchProjections(projections []projection.SearchProjection) {
	if len(projections) == 0 {
		fmt.Println("No results")
		return
	}

	nameWidth := maxProjectionUsernameWidth(projections)
	for i, proj := range projections {
		fmt.Printf("%d. %-*s  %.2f\n", i+1, nameWidth, proj.Username, proj.Score)
	}
}

func printDetailedProjectionList(projections []projection.SearchProjection, startRank int, includeConfidence bool) int {
	nameWidth := maxProjectionUsernameWidth(projections)

	for i, proj := range projections {
		rank := startRank + i
		if includeConfidence {
			fmt.Printf("%d. %-*s  %.2f (%s)\n", rank, nameWidth, proj.Username, proj.Score, proj.Confidence)
		} else {
			fmt.Printf("%d. %-*s  %.2f\n", rank, nameWidth, proj.Username, proj.Score)
		}
		fmt.Println("   Why:")
		for _, reason := range proj.Reasons {
			fmt.Printf("   - %s\n", reason)
		}
		if len(proj.Intelligence) > 0 {
			fmt.Println("   Intelligence:")
			for _, dv := range proj.Intelligence {
				fmt.Printf("   - %s: %s\n", dv.Name, dv.Summary)
			}
		}
		fmt.Println()
	}

	return startRank + len(projections)
}

func maxProjectionUsernameWidth(projections []projection.SearchProjection) int {
	width := 1
	for _, proj := range projections {
		if len(proj.Username) > width {
			width = len(proj.Username)
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

func printInspectProjection(proj projection.InspectProjection, showProvenance bool) {
	fmt.Printf("Profile: %s\n", proj.Username)

	if proj.Metadata != nil {
		fmt.Println()
		fmt.Println("Metadata:")
		if proj.Metadata.Name != "" {
			fmt.Printf("  Name:      %s\n", proj.Metadata.Name)
		}
		if proj.Metadata.Company != "" {
			fmt.Printf("  Company:   %s\n", proj.Metadata.Company)
		}
		if proj.Metadata.Location != "" {
			fmt.Printf("  Location:  %s\n", proj.Metadata.Location)
		}
	}

	if proj.Contributions != nil {
		fmt.Println()
		fmt.Println("Contributions:")
		fmt.Printf("  Total:            %d\n", proj.Contributions.TotalContributions)
		fmt.Printf("  Pull requests:    %d\n", proj.Contributions.TotalPullRequests)
		fmt.Printf("  Issues opened:    %d\n", proj.Contributions.IssuesOpened)
	}

	if proj.Facts != nil {
		fmt.Println()
		fmt.Println("Facts:")
		fmt.Printf("  Total repos:       %d\n", proj.Facts.TotalRepos)
		fmt.Printf("  Original repos:    %d\n", proj.Facts.OriginalRepos)
		fmt.Printf("  Fork repos:        %d\n", proj.Facts.ForkRepos)
		fmt.Printf("  Recently active:   %d\n", proj.Facts.RecentRepos)
		fmt.Printf("  Deep repos:        %d\n", proj.Facts.DeepRepos)
		fmt.Printf("  Valid repos:       %d\n", proj.Facts.ValidRepos)
		fmt.Printf("  Valid original:    %d\n", proj.Facts.ValidOriginalRepos)
		fmt.Printf("  Largest repo:     %d KB\n", proj.Facts.LargestRepoSize)
		fmt.Printf("  Latest activity:  %s\n", proj.Facts.LatestActivity.Format("2006-01-02"))
	}

	if proj.ActivityFacts != nil {
		fmt.Println()
		fmt.Println("Activity:")
		fmt.Printf("  Recent commits:     %d\n", proj.ActivityFacts.RecentCommits)
		fmt.Printf("  Recent PRs:         %d\n", proj.ActivityFacts.RecentPullRequests)
		fmt.Printf("  Recent reviews:     %d\n", proj.ActivityFacts.RecentReviews)
		fmt.Printf("  Recent issues:      %d\n", proj.ActivityFacts.RecentIssues)
		fmt.Printf("  Lifetime total:     %d\n", proj.ActivityFacts.LifetimeTotal)
		fmt.Printf("  Active days:        %d\n", proj.ActivityFacts.ActiveDays)
		fmt.Printf("  Contribution freq:  %.1f\n", proj.ActivityFacts.ContributionFrequency)
		fmt.Printf("  Repository breadth: %d\n", proj.ActivityFacts.RepositoryBreadth)
		fmt.Printf("  Contribution breadth: %d\n", proj.ActivityFacts.ContributionBreadth)
	}

	fmt.Println()
	fmt.Println("Signals:")

	names := make([]string, 0, len(proj.Signals))
	for name := range proj.Signals {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("  %-13s  %.2f\n", name, proj.Signals[name])
	}

	if len(proj.Evidence) > 0 {
		fmt.Println()
		fmt.Println("Evidence:")
		for _, group := range proj.Evidence {
			fmt.Printf("  %s:\n", group.Signal)
			for _, item := range group.Items {
				fmt.Printf("    - %s\n", item.Description)
			}
			printProvenance(group.Provenance, showProvenance, "    ")
		}
	}

	if len(proj.Intelligence) > 0 {
		fmt.Println()
		fmt.Println("Intelligence:")
		for _, dv := range proj.Intelligence {
			fmt.Printf("  %s (%s):\n", dv.Name, dv.Level)
			fmt.Printf("    %s\n", dv.Summary)
			for _, g := range dv.Evidence {
				for _, item := range g.Items {
					fmt.Printf("      - %s\n", item.Description)
				}
				printProvenance(g.Provenance, showProvenance, "      ")
			}
		}
	}

	if len(proj.RepositoryIntelligence) > 0 {
		fmt.Println()
		fmt.Println("Repository Intelligence:")
		for _, rv := range proj.RepositoryIntelligence {
			fmt.Printf("  %s:\n", rv.Repository)
			for _, dv := range rv.Dimensions {
				fmt.Printf("    %s: %s\n", dv.Name, dv.Summary)
				for _, g := range dv.Evidence {
					printProvenance(g.Provenance, showProvenance, "      ")
				}
			}
		}
	}
}

// printProvenance renders a provenance chain in human output only when the user
// asks for it. Default inspect output stays concise; --provenance surfaces the
// explanation graph that JSON always carries.
func printProvenance(chain provenance.Chain, show bool, indent string) {
	if !show || chain.IsEmpty() {
		return
	}
	if len(chain.Repositories) > 0 {
		refs := make([]string, 0, len(chain.Repositories))
		for _, r := range chain.Repositories {
			if r.Dimension != "" {
				refs = append(refs, r.Repository+"."+r.Dimension)
			} else {
				refs = append(refs, r.Repository)
			}
		}
		fmt.Printf("%sprovenance repositories: %s\n", indent, strings.Join(refs, ", "))
	}
	if len(chain.Indicators) > 0 {
		refs := make([]string, 0, len(chain.Indicators))
		for _, i := range chain.Indicators {
			refs = append(refs, i.Signal)
		}
		fmt.Printf("%sprovenance indicators: %s\n", indent, strings.Join(refs, ", "))
	}
	if len(chain.Facts) > 0 {
		refs := make([]string, 0, len(chain.Facts))
		for _, f := range chain.Facts {
			refs = append(refs, f.Name)
		}
		fmt.Printf("%sprovenance facts: %s\n", indent, strings.Join(refs, ", "))
	}
	if len(chain.Observations) > 0 {
		refs := make([]string, 0, len(chain.Observations))
		for _, o := range chain.Observations {
			label := string(o.Kind)
			if o.ID != "" {
				label += "/" + o.ID
			}
			if o.Field != "" {
				label += "." + o.Field
			}
			refs = append(refs, label)
		}
		fmt.Printf("%sprovenance observations: %s\n", indent, strings.Join(refs, ", "))
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
