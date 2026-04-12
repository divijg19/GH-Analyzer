package ghanalyzer

import "testing"

func TestExecuteFilteringAndRanking(t *testing.T) {
	idx := Index{Profiles: []Profile{
		{Username: "alice", Signals: map[string]float64{"consistency": 0.9, "ownership": 0.8, "depth": 0.7}},
		{Username: "bob", Signals: map[string]float64{"consistency": 0.6, "ownership": 0.9, "depth": 0.8}},
		{Username: "cara", Signals: map[string]float64{"consistency": 0.85, "ownership": 0.4, "depth": 0.9}},
	}}

	query := Query{
		Conditions: []Condition{{Signal: "consistency", Operator: ">=", Value: 0.7}},
	}

	results := Execute(idx, query, WeightedRanking{})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Profile.Username != "alice" {
		t.Fatalf("expected first result alice, got %s", results[0].Profile.Username)
	}
	if results[1].Profile.Username != "cara" {
		t.Fatalf("expected second result cara, got %s", results[1].Profile.Username)
	}
}

func TestExecuteTieBreakByUsername(t *testing.T) {
	idx := Index{Profiles: []Profile{
		{Username: "zoe", Signals: map[string]float64{"consistency": 0.5, "ownership": 0.5, "depth": 0.5}},
		{Username: "amy", Signals: map[string]float64{"consistency": 0.5, "ownership": 0.5, "depth": 0.5}},
	}}

	results := Execute(idx, Query{}, WeightedRanking{})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Profile.Username != "amy" || results[1].Profile.Username != "zoe" {
		t.Fatalf("unexpected order: %s, %s", results[0].Profile.Username, results[1].Profile.Username)
	}
}

func TestExecuteLimit(t *testing.T) {
	idx := Index{Profiles: []Profile{
		{Username: "a", Signals: map[string]float64{"consistency": 0.9, "ownership": 0.9, "depth": 0.9}},
		{Username: "b", Signals: map[string]float64{"consistency": 0.8, "ownership": 0.8, "depth": 0.8}},
		{Username: "c", Signals: map[string]float64{"consistency": 0.7, "ownership": 0.7, "depth": 0.7}},
	}}

	results := Execute(idx, Query{Limit: 2}, WeightedRanking{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
