package evaluation

import "testing"

func TestClassifyConfidence(t *testing.T) {
	tests := []struct {
		score    float64
		expected Confidence
	}{
		{0.9, High},
		{0.8, High},
		{0.76, High},
		{0.75, Moderate},
		{0.6, Moderate},
		{0.51, Moderate},
		{0.5, Low},
		{0.3, Low},
		{0.0, Low},
		{-0.1, Low},
		{1.0, High},
		{0.01, Low},
	}

	for _, tt := range tests {
		result := ClassifyConfidence(tt.score)
		if result != tt.expected {
			t.Errorf("ClassifyConfidence(%f) = %q, want %q", tt.score, result, tt.expected)
		}
	}
}

func TestClassifyConfidence_Deterministic(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := ClassifyConfidence(0.76)
		if result != High {
			t.Errorf("iteration %d: ClassifyConfidence(0.76) = %q, want %q", i, result, High)
		}
	}
}

func TestConfidence_String(t *testing.T) {
	if string(High) != "high" {
		t.Errorf("High = %q, want %q", string(High), "high")
	}
	if string(Moderate) != "moderate" {
		t.Errorf("Moderate = %q, want %q", string(Moderate), "moderate")
	}
	if string(Low) != "low" {
		t.Errorf("Low = %q, want %q", string(Low), "low")
	}
}
