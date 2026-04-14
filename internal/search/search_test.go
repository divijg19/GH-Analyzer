package search

import (
	"math"
	"testing"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
)

func TestSearchKeywordMapping(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	for _, result := range results {
		if result.Profile.Signals["depth"] < 0.6 {
			t.Fatalf("result does not satisfy backend mapping: %+v", result.Profile)
		}
	}
}

func TestSearchPresetMapping(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "", Options{Preset: "strong"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for _, result := range results {
		if result.Profile.Signals["consistency"] < 0.7 || result.Profile.Signals["ownership"] < 0.6 {
			t.Fatalf("result does not satisfy strong preset: %+v", result.Profile)
		}
	}
}

func TestSearchExpressionPassthrough(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "consistency >= 0.7 AND depth >= 0.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result, got %d", len(results))
	}
	if results[0].Profile.Username != "alice" {
		t.Fatalf("expected alice, got %q", results[0].Profile.Username)
	}
}

func TestSearchSmallDatasetReturnsRawTopScore(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "consistency >= 0", Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}

	if math.Abs(results[0].Score-0.81) > 1e-9 {
		t.Fatalf("expected top raw score 0.81 for small dataset fallback, got %.8f", results[0].Score)
	}
}

func TestSearchMixedInputWorks(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "backend consistency > 0.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 mixed-input result, got %d", len(results))
	}
	if results[0].Profile.Username != "alice" {
		t.Fatalf("expected alice, got %q", results[0].Profile.Username)
	}
}

func TestSearchUnknownTokensIgnored(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "randomword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 ranked results for unknown token fallback, got %d", len(results))
	}
}

func TestSearchSystemsKeyword(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "systems")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected no systems matches, got %d", len(results))
	}
}

func TestSearchFrontendKeyword(t *testing.T) {
	idx := fixtureIndex()

	results, err := Search(idx, "frontend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 frontend match, got %d", len(results))
	}
	if results[0].Profile.Username != "alice" {
		t.Fatalf("expected alice, got %q", results[0].Profile.Username)
	}
}

func fixtureIndex() indexpkg.Index {
	return indexpkg.Index{Profiles: []indexpkg.Profile{
		{
			Username: "alice",
			Signals: map[string]float64{
				"consistency": 0.9,
				"ownership":   0.8,
				"depth":       0.7,
				"activity":    1.0,
			},
		},
		{
			Username: "bob",
			Signals: map[string]float64{
				"consistency": 0.5,
				"ownership":   0.4,
				"depth":       0.3,
				"activity":    1.0,
			},
		},
	}}
}
