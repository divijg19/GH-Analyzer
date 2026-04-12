package engine

import (
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

func TestMatchOperators(t *testing.T) {
	profile := index.Profile{Signals: map[string]float64{"consistency": 0.7}}

	tests := []struct {
		name string
		cond Condition
		want bool
	}{
		{name: "greater than true", cond: Condition{Signal: "consistency", Operator: ">", Value: 0.6}, want: true},
		{name: "greater than false", cond: Condition{Signal: "consistency", Operator: ">", Value: 0.7}, want: false},
		{name: "greater equal true", cond: Condition{Signal: "consistency", Operator: ">=", Value: 0.7}, want: true},
		{name: "less than true", cond: Condition{Signal: "consistency", Operator: "<", Value: 0.8}, want: true},
		{name: "less than false", cond: Condition{Signal: "consistency", Operator: "<", Value: 0.7}, want: false},
		{name: "less equal true", cond: Condition{Signal: "consistency", Operator: "<=", Value: 0.7}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Match(profile, tc.cond)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMatchEdgeCases(t *testing.T) {
	profile := index.Profile{Signals: map[string]float64{"consistency": 0.7}}

	if Match(profile, Condition{Signal: "ownership", Operator: ">=", Value: 0.5}) {
		t.Fatal("expected false for missing signal")
	}

	if Match(profile, Condition{Signal: "consistency", Operator: "!=", Value: 0.7}) {
		t.Fatal("expected false for invalid operator")
	}
}
