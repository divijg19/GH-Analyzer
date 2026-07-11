package signals

import (
	"fmt"
	"time"
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

func GenerateEvidence(facts RepositoryFacts, signals Signals) []EvidenceGroup {
	groups := make([]EvidenceGroup, 0, 4)

	groups = append(groups, EvidenceGroup{
		Signal: SignalOwnership,
		Items: []Evidence{
			{
				Kind:        "fact",
				Description: fmt.Sprintf("%d of %d repos with content are original", facts.ValidOriginalRepos, facts.ValidRepos),
				Value:       fmt.Sprintf("%d", facts.ValidOriginalRepos),
			},
			{
				Kind:        "signal",
				Description: fmt.Sprintf("Ownership score: %.2f", signals.Ownership),
				Value:       fmt.Sprintf("%.2f", signals.Ownership),
			},
		},
	})

	groups = append(groups, EvidenceGroup{
		Signal: SignalConsistency,
		Items: []Evidence{
			{
				Kind:        "fact",
				Description: fmt.Sprintf("%d of %d original repos active in last 90 days", facts.RecentRepos, facts.OriginalRepos),
				Value:       fmt.Sprintf("%d", facts.RecentRepos),
			},
			{
				Kind:        "signal",
				Description: fmt.Sprintf("Consistency score: %.2f", signals.Consistency),
				Value:       fmt.Sprintf("%.2f", signals.Consistency),
			},
		},
	})

	groups = append(groups, EvidenceGroup{
		Signal: SignalDepth,
		Items: []Evidence{
			{
				Kind:        "fact",
				Description: fmt.Sprintf("%d of %d original repos exceed 50KB", facts.DeepRepos, facts.OriginalRepos),
				Value:       fmt.Sprintf("%d", facts.DeepRepos),
			},
			{
				Kind:        "fact",
				Description: fmt.Sprintf("Largest repo: %dKB", facts.LargestRepoSize),
				Value:       fmt.Sprintf("%d", facts.LargestRepoSize),
			},
			{
				Kind:        "signal",
				Description: fmt.Sprintf("Depth score: %.2f", signals.Depth),
				Value:       fmt.Sprintf("%.2f", signals.Depth),
			},
		},
	})

	groups = append(groups, EvidenceGroup{
		Signal: SignalActivity,
		Items: []Evidence{
			{
				Kind:        "fact",
				Description: fmt.Sprintf("Latest activity: %s", formatActivityDate(facts.LatestActivity)),
				Value:       formatActivityDate(facts.LatestActivity),
			},
			{
				Kind:        "signal",
				Description: fmt.Sprintf("Activity score: %.2f", signals.Activity),
				Value:       fmt.Sprintf("%.2f", signals.Activity),
			},
		},
	})

	return groups
}

func formatActivityDate(t time.Time) string {
	if t.IsZero() {
		return "none"
	}
	return t.Format("2006-01-02")
}
