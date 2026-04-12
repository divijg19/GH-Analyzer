package ghanalyzer

import "testing"

func TestExplainIncludesValuesAndDeduplicates(t *testing.T) {
	profile := Profile{Signals: map[string]float64{
		"consistency": 0.87,
		"ownership":   0.75,
	}}

	query := Query{Conditions: []Condition{
		{Signal: "consistency", Operator: ">=", Value: 0.7},
		{Signal: "consistency", Operator: ">", Value: 0.5},
		{Signal: "ownership", Operator: ">=", Value: 0.6},
	}}

	reasons := Explain(profile, query)
	if len(reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(reasons))
	}

	if reasons[0] != "High consistency (0.87)" {
		t.Fatalf("unexpected first reason: %s", reasons[0])
	}
	if reasons[1] != "Strong ownership (0.75)" {
		t.Fatalf("unexpected second reason: %s", reasons[1])
	}
}

func TestExplainNoConditions(t *testing.T) {
	reasons := Explain(Profile{}, Query{})
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d", len(reasons))
	}
	if reasons[0] != "Ranked by overall signal strength" {
		t.Fatalf("unexpected reason: %s", reasons[0])
	}
}
