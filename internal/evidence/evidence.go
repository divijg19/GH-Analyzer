package evidence

import (
	"fmt"
	"sort"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/indicators"
	"github.com/divijg19/Atlas/internal/provenance"
)

type Evidence struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Value       string `json:"value"`
}

// EvidenceGroup is the canonical unit of explanation in Atlas. Its items render
// the human-readable reasons for a conclusion; its Provenance chain binds those
// reasons to the indicators, facts and observations that produced them. Evidence
// owns explanation for every layer above it.
type EvidenceGroup struct {
	Signal     string           `json:"signal"`
	Items      []Evidence       `json:"items"`
	Provenance provenance.Chain `json:"provenance,omitempty"`
}

// PortfolioGroup is the evidence-group label for deterministic Portfolio
// Intelligence facts. It is NOT an indicator signal: signals are reserved for the
// four scored indicator families (ownership, consistency, depth, activity). The
// portfolio group records what was observed/aggregated, not an interpreted score.
const PortfolioGroup = "portfolio"

func GenerateEvidence(factsValue facts.RepositoryFacts, s indicators.Signals) []EvidenceGroup {
	groups := make([]EvidenceGroup, 0, 4)
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalOwnership, Provenance: indicators.SignalProvenance(indicators.SignalOwnership), Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d repos with content are original", factsValue.ValidOriginalRepos, factsValue.ValidRepos), Value: fmt.Sprintf("%d", factsValue.ValidOriginalRepos)}, {Kind: "signal", Description: fmt.Sprintf("Ownership score: %.2f", s.Ownership), Value: fmt.Sprintf("%.2f", s.Ownership)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalConsistency, Provenance: indicators.SignalProvenance(indicators.SignalConsistency), Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d original repos active in last 90 days", factsValue.RecentRepos, factsValue.OriginalRepos), Value: fmt.Sprintf("%d", factsValue.RecentRepos)}, {Kind: "signal", Description: fmt.Sprintf("Consistency score: %.2f", s.Consistency), Value: fmt.Sprintf("%.2f", s.Consistency)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalDepth, Provenance: indicators.SignalProvenance(indicators.SignalDepth), Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d original repos exceed 50KB", factsValue.DeepRepos, factsValue.OriginalRepos), Value: fmt.Sprintf("%d", factsValue.DeepRepos)}, {Kind: "fact", Description: fmt.Sprintf("Largest repo: %dKB", factsValue.LargestRepoSize), Value: fmt.Sprintf("%d", factsValue.LargestRepoSize)}, {Kind: "signal", Description: fmt.Sprintf("Depth score: %.2f", s.Depth), Value: fmt.Sprintf("%.2f", s.Depth)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalActivity, Provenance: indicators.SignalProvenance(indicators.SignalActivity), Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("Latest activity: %s", formatActivityDate(factsValue.LatestActivity)), Value: formatActivityDate(factsValue.LatestActivity)}, {Kind: "signal", Description: fmt.Sprintf("Activity score: %.2f", s.Activity), Value: fmt.Sprintf("%.2f", s.Activity)}}})
	groups = append(groups, EvidenceGroup{Signal: PortfolioGroup, Provenance: facts.FactsProvenance("ranked_topics", "fork_lineage", "technology_timeline", "maintenance_buckets"), Items: portfolioEvidenceItems(factsValue)})
	return groups
}

func formatActivityDate(t time.Time) string {
	if t.IsZero() {
		return "none"
	}
	return t.Format("2006-01-02")
}

// portfolioEvidenceItems renders the deterministic Portfolio Intelligence facts as
// explanation items. These are aggregations only; no score or interpretation is
// attached.
func portfolioEvidenceItems(f facts.RepositoryFacts) []Evidence {
	items := []Evidence{
		{Kind: "fact", Description: fmt.Sprintf("%d distinct topics across portfolio", f.TopicUniverse), Value: fmt.Sprintf("%d", f.TopicUniverse)},
	}
	if len(f.RankedTopics) > 0 {
		items = append(items, Evidence{Kind: "fact", Description: "Top topics", Value: joinTopics(f.RankedTopics)})
	}
	if len(f.ForkLineage) > 0 {
		concentration := f.ForkLineage[0]
		items = append(items, Evidence{Kind: "fact", Description: fmt.Sprintf("Most-forked upstream: %s", concentration.Parent), Value: fmt.Sprintf("%d forks", concentration.Forks)})
	}
	if len(f.TechnologyTimeline) > 0 {
		items = append(items, Evidence{Kind: "fact", Description: "Technology timeline (year: languages)", Value: formatEras(f.TechnologyTimeline)})
	}
	items = append(items, Evidence{Kind: "fact", Description: fmt.Sprintf("Maintenance buckets (active/recent/dormant): %d/%d/%d", f.MaintenanceBuckets.Active, f.MaintenanceBuckets.Recent, f.MaintenanceBuckets.Dormant), Value: fmt.Sprintf("%d/%d/%d", f.MaintenanceBuckets.Active, f.MaintenanceBuckets.Recent, f.MaintenanceBuckets.Dormant)})
	return items
}

func joinTopics(topics []string) string {
	out := ""
	for i, t := range topics {
		if i > 0 {
			out += ", "
		}
		out += t
	}
	return out
}

func formatEras(eras map[int][]string) string {
	years := make([]int, 0, len(eras))
	for y := range eras {
		years = append(years, y)
	}
	sortInts(years)
	out := ""
	for i, y := range years {
		if i > 0 {
			out += "; "
		}
		out += fmt.Sprintf("%d: %s", y, joinTopics(eras[y]))
	}
	return out
}

func sortInts(s []int) {
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
}
