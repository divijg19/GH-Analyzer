package evidence

import (
	"fmt"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/indicators"
)

type Evidence struct {
	Kind        string
	Description string
	Value       string
}

type EvidenceGroup struct {
	Signal string
	Items  []Evidence
}

func GenerateEvidence(factsValue facts.RepositoryFacts, s indicators.Signals) []EvidenceGroup {
	groups := make([]EvidenceGroup, 0, 4)
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalOwnership, Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d repos with content are original", factsValue.ValidOriginalRepos, factsValue.ValidRepos), Value: fmt.Sprintf("%d", factsValue.ValidOriginalRepos)}, {Kind: "signal", Description: fmt.Sprintf("Ownership score: %.2f", s.Ownership), Value: fmt.Sprintf("%.2f", s.Ownership)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalConsistency, Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d original repos active in last 90 days", factsValue.RecentRepos, factsValue.OriginalRepos), Value: fmt.Sprintf("%d", factsValue.RecentRepos)}, {Kind: "signal", Description: fmt.Sprintf("Consistency score: %.2f", s.Consistency), Value: fmt.Sprintf("%.2f", s.Consistency)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalDepth, Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("%d of %d original repos exceed 50KB", factsValue.DeepRepos, factsValue.OriginalRepos), Value: fmt.Sprintf("%d", factsValue.DeepRepos)}, {Kind: "fact", Description: fmt.Sprintf("Largest repo: %dKB", factsValue.LargestRepoSize), Value: fmt.Sprintf("%d", factsValue.LargestRepoSize)}, {Kind: "signal", Description: fmt.Sprintf("Depth score: %.2f", s.Depth), Value: fmt.Sprintf("%.2f", s.Depth)}}})
	groups = append(groups, EvidenceGroup{Signal: indicators.SignalActivity, Items: []Evidence{{Kind: "fact", Description: fmt.Sprintf("Latest activity: %s", formatActivityDate(factsValue.LatestActivity)), Value: formatActivityDate(factsValue.LatestActivity)}, {Kind: "signal", Description: fmt.Sprintf("Activity score: %.2f", s.Activity), Value: fmt.Sprintf("%.2f", s.Activity)}}})
	return groups
}

func formatActivityDate(t time.Time) string {
	if t.IsZero() {
		return "none"
	}
	return t.Format("2006-01-02")
}
