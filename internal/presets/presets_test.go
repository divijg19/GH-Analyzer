package presets

import (
	"testing"

	"github.com/divijg19/Atlas/internal/engine"
)

func TestPreset(t *testing.T) {
	tests := []struct {
		name       string
		presetName string
		want       []engine.Condition
		wantErr    bool
	}{
		{
			name:       "strong",
			presetName: "strong",
			want: []engine.Condition{
				{Signal: "consistency", Operator: ">=", Value: 0.7},
				{Signal: "ownership", Operator: ">=", Value: 0.6},
			},
		},
		{
			name:       "consistent",
			presetName: "consistent",
			want: []engine.Condition{
				{Signal: "consistency", Operator: ">=", Value: 0.8},
			},
		},
		{
			name:       "deep",
			presetName: "deep",
			want: []engine.Condition{
				{Signal: "depth", Operator: ">=", Value: 0.7},
			},
		},
		{name: "invalid", presetName: "unknown", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := Preset(tc.presetName)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(query.Conditions) != len(tc.want) {
				t.Fatalf("expected %d conditions, got %d", len(tc.want), len(query.Conditions))
			}

			for i := range tc.want {
				if query.Conditions[i] != tc.want[i] {
					t.Fatalf("condition %d mismatch: expected %+v, got %+v", i, tc.want[i], query.Conditions[i])
				}
			}
		})
	}
}
