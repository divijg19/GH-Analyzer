package signals

import (
	"testing"
	"time"
)

func TestGenerateEvidenceDeterministic(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	facts := RepositoryFacts{
		TotalRepos:         42,
		OriginalRepos:      38,
		ForkRepos:          4,
		RecentRepos:        35,
		DeepRepos:          12,
		ValidRepos:         39,
		ValidOriginalRepos: 35,
		LargestRepoSize:    2048,
		LatestActivity:     now.Add(-72 * time.Hour),
	}
	signals := Signals{
		Ownership:   0.90,
		Consistency: 0.85,
		Depth:       0.60,
		Activity:    1.00,
	}

	first := GenerateEvidence(facts, signals)
	second := GenerateEvidence(facts, signals)

	if len(first) != len(second) {
		t.Fatalf("determinism: group count changed between calls")
	}

	for i := range first {
		if first[i].Signal != second[i].Signal {
			t.Fatalf("determinism: signal at index %d changed: %q vs %q", i, first[i].Signal, second[i].Signal)
		}
		if len(first[i].Items) != len(second[i].Items) {
			t.Fatalf("determinism: item count for %s changed", first[i].Signal)
		}
		for j := range first[i].Items {
			if first[i].Items[j].Kind != second[i].Items[j].Kind {
				t.Fatalf("determinism: %s[%d].Kind changed", first[i].Signal, j)
			}
			if first[i].Items[j].Description != second[i].Items[j].Description {
				t.Fatalf("determinism: %s[%d].Description changed", first[i].Signal, j)
			}
			if first[i].Items[j].Value != second[i].Items[j].Value {
				t.Fatalf("determinism: %s[%d].Value changed", first[i].Signal, j)
			}
		}
	}
}

func TestGenerateEvidenceStructure(t *testing.T) {
	facts := RepositoryFacts{
		OriginalRepos:      10,
		RecentRepos:        8,
		DeepRepos:          5,
		ValidRepos:         12,
		ValidOriginalRepos: 9,
		LargestRepoSize:    500,
		LatestActivity:     time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	signals := Signals{
		Ownership:   0.75,
		Consistency: 0.80,
		Depth:       0.50,
		Activity:    0.70,
	}

	groups := GenerateEvidence(facts, signals)

	if len(groups) != 4 {
		t.Fatalf("expected 4 evidence groups, got %d", len(groups))
	}

	signalNames := []string{"ownership", "consistency", "depth", "activity"}
	for i, name := range signalNames {
		if groups[i].Signal != name {
			t.Fatalf("group[%d].Signal = %q, want %q", i, groups[i].Signal, name)
		}
	}
}

func TestGenerateEvidenceOwnershipGroup(t *testing.T) {
	facts := RepositoryFacts{
		ValidRepos:         39,
		ValidOriginalRepos: 35,
	}
	signals := Signals{Ownership: 0.90}

	groups := GenerateEvidence(facts, signals)
	group := findGroup(groups, "ownership")

	if group == nil {
		t.Fatal("ownership group not found")
	}
	if len(group.Items) != 2 {
		t.Fatalf("ownership: expected 2 items, got %d", len(group.Items))
	}

	if group.Items[0].Kind != "fact" {
		t.Fatalf("ownership item[0].Kind = %q, want %q", group.Items[0].Kind, "fact")
	}
	if group.Items[0].Description != "35 of 39 repos with content are original" {
		t.Fatalf("ownership fact description = %q", group.Items[0].Description)
	}
	if group.Items[0].Value != "35" {
		t.Fatalf("ownership fact value = %q, want %q", group.Items[0].Value, "35")
	}

	if group.Items[1].Kind != "signal" {
		t.Fatalf("ownership item[1].Kind = %q, want %q", group.Items[1].Kind, "signal")
	}
	if group.Items[1].Description != "Ownership score: 0.90" {
		t.Fatalf("ownership signal description = %q", group.Items[1].Description)
	}
	if group.Items[1].Value != "0.90" {
		t.Fatalf("ownership signal value = %q, want %q", group.Items[1].Value, "0.90")
	}
}

func TestGenerateEvidenceConsistencyGroup(t *testing.T) {
	facts := RepositoryFacts{
		OriginalRepos: 38,
		RecentRepos:   35,
	}
	signals := Signals{Consistency: 0.85}

	groups := GenerateEvidence(facts, signals)
	group := findGroup(groups, "consistency")

	if group == nil {
		t.Fatal("consistency group not found")
	}
	if len(group.Items) != 2 {
		t.Fatalf("consistency: expected 2 items, got %d", len(group.Items))
	}

	if group.Items[0].Kind != "fact" {
		t.Fatalf("consistency item[0].Kind = %q, want fact", group.Items[0].Kind)
	}
	if group.Items[0].Description != "35 of 38 original repos active in last 90 days" {
		t.Fatalf("consistency fact description = %q", group.Items[0].Description)
	}
	if group.Items[0].Value != "35" {
		t.Fatalf("consistency fact value = %q, want %q", group.Items[0].Value, "35")
	}

	if group.Items[1].Description != "Consistency score: 0.85" {
		t.Fatalf("consistency signal description = %q", group.Items[1].Description)
	}
	if group.Items[1].Value != "0.85" {
		t.Fatalf("consistency signal value = %q, want %q", group.Items[1].Value, "0.85")
	}
}

func TestGenerateEvidenceDepthGroup(t *testing.T) {
	facts := RepositoryFacts{
		OriginalRepos:   38,
		DeepRepos:       12,
		LargestRepoSize: 2048,
	}
	signals := Signals{Depth: 0.60}

	groups := GenerateEvidence(facts, signals)
	group := findGroup(groups, "depth")

	if group == nil {
		t.Fatal("depth group not found")
	}
	if len(group.Items) != 3 {
		t.Fatalf("depth: expected 3 items, got %d", len(group.Items))
	}

	if group.Items[0].Kind != "fact" {
		t.Fatalf("depth item[0].Kind = %q, want fact", group.Items[0].Kind)
	}
	if group.Items[0].Description != "12 of 38 original repos exceed 50KB" {
		t.Fatalf("depth fact1 description = %q", group.Items[0].Description)
	}
	if group.Items[0].Value != "12" {
		t.Fatalf("depth fact1 value = %q, want %q", group.Items[0].Value, "12")
	}

	if group.Items[1].Kind != "fact" {
		t.Fatalf("depth item[1].Kind = %q, want fact", group.Items[1].Kind)
	}
	if group.Items[1].Description != "Largest repo: 2048KB" {
		t.Fatalf("depth fact2 description = %q", group.Items[1].Description)
	}
	if group.Items[1].Value != "2048" {
		t.Fatalf("depth fact2 value = %q, want %q", group.Items[1].Value, "2048")
	}

	if group.Items[2].Kind != "signal" {
		t.Fatalf("depth item[2].Kind = %q, want signal", group.Items[2].Kind)
	}
	if group.Items[2].Description != "Depth score: 0.60" {
		t.Fatalf("depth signal description = %q", group.Items[2].Description)
	}
	if group.Items[2].Value != "0.60" {
		t.Fatalf("depth signal value = %q, want %q", group.Items[2].Value, "0.60")
	}
}

func TestGenerateEvidenceActivityGroup(t *testing.T) {
	activityDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	facts := RepositoryFacts{
		LatestActivity: activityDate,
	}
	signals := Signals{Activity: 1.00}

	groups := GenerateEvidence(facts, signals)
	group := findGroup(groups, "activity")

	if group == nil {
		t.Fatal("activity group not found")
	}
	if len(group.Items) != 2 {
		t.Fatalf("activity: expected 2 items, got %d", len(group.Items))
	}

	if group.Items[0].Kind != "fact" {
		t.Fatalf("activity item[0].Kind = %q, want fact", group.Items[0].Kind)
	}
	if group.Items[0].Description != "Latest activity: 2026-05-15" {
		t.Fatalf("activity fact description = %q", group.Items[0].Description)
	}
	if group.Items[0].Value != "2026-05-15" {
		t.Fatalf("activity fact value = %q, want %q", group.Items[0].Value, "2026-05-15")
	}

	if group.Items[1].Description != "Activity score: 1.00" {
		t.Fatalf("activity signal description = %q", group.Items[1].Description)
	}
	if group.Items[1].Value != "1.00" {
		t.Fatalf("activity signal value = %q, want %q", group.Items[1].Value, "1.00")
	}
}

func TestGenerateEvidenceEmptyFacts(t *testing.T) {
	facts := RepositoryFacts{}
	signals := Signals{}

	groups := GenerateEvidence(facts, signals)

	if len(groups) != 4 {
		t.Fatalf("empty: expected 4 groups, got %d", len(groups))
	}

	ownership := findGroup(groups, "ownership")
	if ownership.Items[0].Description != "0 of 0 repos with content are original" {
		t.Fatalf("empty ownership fact: %q", ownership.Items[0].Description)
	}

	consistency := findGroup(groups, "consistency")
	if consistency.Items[0].Description != "0 of 0 original repos active in last 90 days" {
		t.Fatalf("empty consistency fact: %q", consistency.Items[0].Description)
	}

	depth := findGroup(groups, "depth")
	if depth.Items[0].Description != "0 of 0 original repos exceed 50KB" {
		t.Fatalf("empty depth fact1: %q", depth.Items[0].Description)
	}
	if depth.Items[1].Description != "Largest repo: 0KB" {
		t.Fatalf("empty depth fact2: %q", depth.Items[1].Description)
	}

	activity := findGroup(groups, "activity")
	if activity.Items[0].Description != "Latest activity: none" {
		t.Fatalf("empty activity fact: %q", activity.Items[0].Description)
	}
}

func TestGenerateEvidenceZeroTimeActivity(t *testing.T) {
	facts := RepositoryFacts{
		OriginalRepos: 5,
		RecentRepos:   3,
		DeepRepos:     2,
		ValidRepos:    4,
	}
	signals := Signals{Activity: 0.0}

	groups := GenerateEvidence(facts, signals)
	activity := findGroup(groups, "activity")

	if activity.Items[0].Description != "Latest activity: none" {
		t.Fatalf("zero time activity fact: %q", activity.Items[0].Description)
	}
	if activity.Items[0].Value != "none" {
		t.Fatalf("zero time activity value: %q", activity.Items[0].Value)
	}
}

func TestGenerateEvidenceZeroOriginalRepos(t *testing.T) {
	facts := RepositoryFacts{
		OriginalRepos:      0,
		RecentRepos:        0,
		DeepRepos:          0,
		ValidRepos:         5,
		ValidOriginalRepos: 0,
		LargestRepoSize:    100,
		LatestActivity:     time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	signals := Signals{
		Ownership:   0.0,
		Consistency: 0.0,
		Depth:       0.0,
		Activity:    1.0,
	}

	groups := GenerateEvidence(facts, signals)

	ownership := findGroup(groups, "ownership")
	if ownership.Items[0].Description != "0 of 5 repos with content are original" {
		t.Fatalf("zero original ownership fact: %q", ownership.Items[0].Description)
	}

	consistency := findGroup(groups, "consistency")
	if consistency.Items[0].Description != "0 of 0 original repos active in last 90 days" {
		t.Fatalf("zero original consistency fact: %q", consistency.Items[0].Description)
	}

	depth := findGroup(groups, "depth")
	if depth.Items[0].Description != "0 of 0 original repos exceed 50KB" {
		t.Fatalf("zero original depth fact: %q", depth.Items[0].Description)
	}
}

func findGroup(groups []EvidenceGroup, signal string) *EvidenceGroup {
	for i := range groups {
		if groups[i].Signal == signal {
			return &groups[i]
		}
	}
	return nil
}
