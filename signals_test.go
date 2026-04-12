package ghanalyzer

import "testing"

func TestSignalsFromReportNormalization(t *testing.T) {
	report := Report{
		Scores: Scores{
			Consistency: 120,
			Ownership:   -10,
			Depth:       50,
		},
	}

	signals := SignalsFromReport(report)

	if signals["consistency"] != 1.0 {
		t.Fatalf("expected consistency 1.0, got %.2f", signals["consistency"])
	}
	if signals["ownership"] != 0.0 {
		t.Fatalf("expected ownership 0.0, got %.2f", signals["ownership"])
	}
	if signals["depth"] != 0.5 {
		t.Fatalf("expected depth 0.5, got %.2f", signals["depth"])
	}
}

func TestSignalsFromReportActivityDetection(t *testing.T) {
	tests := []struct {
		name   string
		report Report
		want   float64
	}{
		{
			name: "active when consistency above zero",
			report: Report{
				Scores: Scores{Consistency: 10},
			},
			want: 1.0,
		},
		{
			name: "inactive when consistency is zero",
			report: Report{
				Scores:     Scores{Consistency: 0},
				Highlights: []string{"Active in last 90 days"},
			},
			want: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signals := SignalsFromReport(tc.report)
			if signals["activity"] != tc.want {
				t.Fatalf("expected activity %.1f, got %.1f", tc.want, signals["activity"])
			}
		})
	}
}
