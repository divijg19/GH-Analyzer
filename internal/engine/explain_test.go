package engine

import (
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/index"
)

func TestExplainIncludesValuesAndDeduplicates(t *testing.T) {
	profile := index.Profile{Signals: map[string]float64{
		"consistency": 0.87,
		"ownership":   0.75,
	}}

	query := Query{Conditions: []Condition{
		{Signal: "consistency", Operator: ">=", Value: 0.7},
		{Signal: "consistency", Operator: ">", Value: 0.5},
		{Signal: "ownership", Operator: ">=", Value: 0.6},
	}}

	reasons := explain(profile, query)
	if len(reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(reasons))
	}

	if reasons[0] != "High consistency (0.87 >= 0.70)" {
		t.Fatalf("unexpected first reason: %s", reasons[0])
	}
	if reasons[1] != "Strong ownership (0.75 >= 0.60)" {
		t.Fatalf("unexpected second reason: %s", reasons[1])
	}
}

func TestExplainNoConditions(t *testing.T) {
	reasons := explain(index.Profile{}, Query{})
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d", len(reasons))
	}
	if reasons[0] != "Ranked by overall signal strength" {
		t.Fatalf("unexpected reason: %s", reasons[0])
	}
}

func TestExplainWithEvidence(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	facts := &facts.RepositoryFacts{
		OriginalRepos:      38,
		RecentRepos:        35,
		DeepRepos:          12,
		ValidRepos:         39,
		ValidOriginalRepos: 35,
		LargestRepoSize:    2048,
		LatestActivity:     now.Add(-72 * time.Hour),
	}
	profile := index.Profile{
		Facts: facts,
		Signals: map[string]float64{
			"ownership":   0.90,
			"consistency": 0.85,
			"depth":       0.60,
			"activity":    1.00,
		},
	}

	query := Query{Conditions: []Condition{
		{Signal: "ownership", Operator: ">=", Value: 0.6},
		{Signal: "consistency", Operator: ">=", Value: 0.7},
	}}

	reasons := explain(profile, query)

	if len(reasons) == 0 {
		t.Fatal("expected reasons, got none")
	}

	hasOwnershipFact := false
	hasOwnershipSignal := false
	hasConsistencyFact := false
	hasConsistencySignal := false

	for _, r := range reasons {
		if r == "35 of 39 repos with content are original" {
			hasOwnershipFact = true
		}
		if r == "Ownership score: 0.90" {
			hasOwnershipSignal = true
		}
		if r == "35 of 38 original repos active in last 90 days" {
			hasConsistencyFact = true
		}
		if r == "Consistency score: 0.85" {
			hasConsistencySignal = true
		}
	}

	if !hasOwnershipFact {
		t.Fatal("missing ownership fact evidence")
	}
	if !hasOwnershipSignal {
		t.Fatal("missing ownership signal evidence")
	}
	if !hasConsistencyFact {
		t.Fatal("missing consistency fact evidence")
	}
	if !hasConsistencySignal {
		t.Fatal("missing consistency signal evidence")
	}
}

func TestExplainWithEvidenceOnlyMatchedSignals(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	facts := &facts.RepositoryFacts{
		OriginalRepos:      38,
		RecentRepos:        35,
		DeepRepos:          12,
		ValidRepos:         39,
		ValidOriginalRepos: 35,
		LargestRepoSize:    2048,
		LatestActivity:     now.Add(-72 * time.Hour),
	}
	profile := index.Profile{
		Facts: facts,
		Signals: map[string]float64{
			"ownership":   0.90,
			"consistency": 0.85,
			"depth":       0.60,
			"activity":    1.00,
		},
	}

	query := Query{Conditions: []Condition{
		{Signal: "ownership", Operator: ">=", Value: 0.6},
	}}

	reasons := explain(profile, query)

	for _, r := range reasons {
		if r == "35 of 38 original repos active in last 90 days" {
			t.Fatal("consistency evidence should not appear when only ownership matched")
		}
		if r == "Consistency score: 0.85" {
			t.Fatal("consistency signal should not appear when only ownership matched")
		}
		if r == "12 of 38 original repos exceed 50KB" {
			t.Fatal("depth evidence should not appear when only ownership matched")
		}
		if r == "Latest activity: 2026-06-01" {
			t.Fatal("activity evidence should not appear when only ownership matched")
		}
	}

	hasOwnershipFact := false
	for _, r := range reasons {
		if r == "35 of 39 repos with content are original" {
			hasOwnershipFact = true
			break
		}
	}
	if !hasOwnershipFact {
		t.Fatal("ownership evidence should appear")
	}
}

func TestExplainBackwardCompatibleWithoutFacts(t *testing.T) {
	profile := index.Profile{Signals: map[string]float64{
		"consistency": 0.87,
		"ownership":   0.75,
	}}

	query := Query{Conditions: []Condition{
		{Signal: "consistency", Operator: ">=", Value: 0.7},
	}}

	reasons := explain(profile, query)

	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d", len(reasons))
	}
	if reasons[0] != "High consistency (0.87 >= 0.70)" {
		t.Fatalf("expected fallback reason format, got %q", reasons[0])
	}
}
