package projection

import (
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/evaluation"
	"github.com/divijg19/GH-Analyzer/internal/index"
)

func TestBuildSearchProjection_DeterministicOutput(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
		Signals: map[string]float64{
			"consistency": 0.8,
			"ownership":   0.7,
		},
	}
	score := 0.85
	confidence := evaluation.High
	reasons := []string{"high consistency", "strong ownership"}

	for i := 0; i < 10; i++ {
		proj := BuildSearchProjection(profile, score, confidence, reasons)

		if proj.Username != "testuser" {
			t.Fatalf("iteration %d: unexpected Username", i)
		}
		if proj.Score != 0.85 {
			t.Fatalf("iteration %d: unexpected Score", i)
		}
		if proj.Confidence != evaluation.High {
			t.Fatalf("iteration %d: unexpected Confidence", i)
		}
		if len(proj.Reasons) != 2 {
			t.Fatalf("iteration %d: expected 2 reasons", i)
		}
	}
}

func TestBuildSearchProjection_ConfidencePreserved(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
	}

	tests := []struct {
		score    float64
		conf     evaluation.Confidence
		expected evaluation.Confidence
	}{
		{0.9, evaluation.High, evaluation.High},
		{0.6, evaluation.Moderate, evaluation.Moderate},
		{0.3, evaluation.Low, evaluation.Low},
	}

	for _, tt := range tests {
		proj := BuildSearchProjection(profile, tt.score, tt.conf, nil)
		if proj.Confidence != tt.expected {
			t.Errorf("score %.2f: expected confidence %q, got %q", tt.score, tt.expected, proj.Confidence)
		}
	}
}

func TestBuildSearchProjection_SignalPreservation(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
		Signals: map[string]float64{
			"consistency": 0.8,
			"ownership":   0.7,
			"depth":       0.6,
			"activity":    0.5,
		},
	}

	proj := BuildSearchProjection(profile, 0.75, evaluation.Moderate, nil)

	if proj.Signals == nil {
		t.Fatal("expected Signals")
	}
	if len(proj.Signals) != 4 {
		t.Fatalf("expected 4 signals, got %d", len(proj.Signals))
	}
	if proj.Signals["consistency"] != 0.8 {
		t.Fatalf("expected consistency 0.8, got %f", proj.Signals["consistency"])
	}
	if proj.Signals["ownership"] != 0.7 {
		t.Fatalf("expected ownership 0.7, got %f", proj.Signals["ownership"])
	}
	if proj.Signals["depth"] != 0.6 {
		t.Fatalf("expected depth 0.6, got %f", proj.Signals["depth"])
	}
	if proj.Signals["activity"] != 0.5 {
		t.Fatalf("expected activity 0.5, got %f", proj.Signals["activity"])
	}
}

func TestBuildSearchProjection_ReasonsPreservation(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
	}
	reasons := []string{"high consistency", "strong ownership", "deep repos"}

	proj := BuildSearchProjection(profile, 0.8, evaluation.High, reasons)

	if proj.Reasons == nil {
		t.Fatal("expected Reasons")
	}
	if len(proj.Reasons) != 3 {
		t.Fatalf("expected 3 reasons, got %d", len(proj.Reasons))
	}
	if proj.Reasons[0] != "high consistency" {
		t.Fatalf("expected first reason 'high consistency', got %q", proj.Reasons[0])
	}
	if proj.Reasons[1] != "strong ownership" {
		t.Fatalf("expected second reason 'strong ownership', got %q", proj.Reasons[1])
	}
	if proj.Reasons[2] != "deep repos" {
		t.Fatalf("expected third reason 'deep repos', got %q", proj.Reasons[2])
	}
}

func TestBuildSearchProjection_EmptySignals(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
		Signals:  map[string]float64{},
	}

	proj := BuildSearchProjection(profile, 0.5, evaluation.Moderate, nil)

	if proj.Signals == nil {
		t.Fatal("expected Signals to be non-nil")
	}
	if len(proj.Signals) != 0 {
		t.Fatalf("expected empty Signals, got %d", len(proj.Signals))
	}
}

func TestBuildSearchProjection_NilSignals(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
		Signals:  nil,
	}

	proj := BuildSearchProjection(profile, 0.5, evaluation.Moderate, nil)

	if proj.Signals == nil {
		t.Fatal("expected Signals to be non-nil")
	}
	if len(proj.Signals) != 0 {
		t.Fatalf("expected empty Signals, got %d", len(proj.Signals))
	}
}

func TestBuildSearchProjection_EmptyReasons(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
	}

	proj := BuildSearchProjection(profile, 0.6, evaluation.Moderate, []string{})

	if proj.Reasons == nil {
		t.Fatal("expected Reasons to be non-nil")
	}
	if len(proj.Reasons) != 0 {
		t.Fatalf("expected empty Reasons, got %d", len(proj.Reasons))
	}
}

func TestBuildSearchProjection_NilReasons(t *testing.T) {
	profile := index.Profile{
		Username: "testuser",
	}

	proj := BuildSearchProjection(profile, 0.6, evaluation.Moderate, nil)

	if proj.Reasons == nil {
		t.Fatal("expected Reasons to be non-nil")
	}
	if len(proj.Reasons) != 0 {
		t.Fatalf("expected empty Reasons, got %d", len(proj.Reasons))
	}
}

func TestBuildSearchProjection_ClonedSignals(t *testing.T) {
	originalSignals := map[string]float64{"consistency": 0.8}
	profile := index.Profile{
		Username: "testuser",
		Signals:  originalSignals,
	}

	proj := BuildSearchProjection(profile, 0.7, evaluation.Moderate, nil)

	proj.Signals["consistency"] = 0.9

	if originalSignals["consistency"] != 0.8 {
		t.Fatal("expected Signals to be cloned, not reference")
	}
}

func TestBuildSearchProjection_ClonedReasons(t *testing.T) {
	originalReasons := []string{"reason1", "reason2"}
	profile := index.Profile{
		Username: "testuser",
	}

	proj := BuildSearchProjection(profile, 0.7, evaluation.Moderate, originalReasons)

	proj.Reasons[0] = "modified"

	if originalReasons[0] != "reason1" {
		t.Fatal("expected Reasons to be cloned, not reference")
	}
}

func TestBuildSearchProjection_FullIntegration(t *testing.T) {
	profile := index.Profile{
		Username: "fulluser",
		Signals: map[string]float64{
			"consistency": 0.8,
			"ownership":   0.7,
			"depth":       0.6,
			"activity":    0.5,
		},
	}
	score := 0.85
	confidence := evaluation.High
	reasons := []string{"high consistency", "strong ownership"}

	proj := BuildSearchProjection(profile, score, confidence, reasons)

	if proj.Username != "fulluser" {
		t.Fatalf("unexpected Username: %s", proj.Username)
	}
	if proj.Score != 0.85 {
		t.Fatalf("unexpected Score: %f", proj.Score)
	}
	if proj.Confidence != evaluation.High {
		t.Fatalf("unexpected Confidence: %s", proj.Confidence)
	}
	if len(proj.Signals) != 4 {
		t.Fatalf("expected 4 signals, got %d", len(proj.Signals))
	}
	if len(proj.Reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(proj.Reasons))
	}
}

func TestCloneSignals(t *testing.T) {
	original := map[string]float64{
		"consistency": 0.8,
		"ownership":   0.7,
	}

	cloned := cloneSignals(original)

	if len(cloned) != len(original) {
		t.Fatalf("expected %d signals, got %d", len(original), len(cloned))
	}
	if cloned["consistency"] != 0.8 {
		t.Fatalf("expected consistency 0.8, got %f", cloned["consistency"])
	}
	if cloned["ownership"] != 0.7 {
		t.Fatalf("expected ownership 0.7, got %f", cloned["ownership"])
	}

	cloned["consistency"] = 0.9
	if original["consistency"] != 0.8 {
		t.Fatal("expected original to be unchanged")
	}
}

func TestCloneSignals_Empty(t *testing.T) {
	original := map[string]float64{}

	cloned := cloneSignals(original)

	if len(cloned) != 0 {
		t.Fatalf("expected empty map, got %d", len(cloned))
	}
}

func TestCloneSignals_Nil(t *testing.T) {
	cloned := cloneSignals(nil)

	if len(cloned) != 0 {
		t.Fatalf("expected empty map, got %d", len(cloned))
	}
}
