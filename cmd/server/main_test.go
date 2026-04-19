package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
)

func TestSearchHandlerDatasetContract(t *testing.T) {
	restore := overrideServerSearchDependencies(
		func(path string) (indexpkg.Index, error) {
			return indexpkg.Index{}, nil
		},
		buildLiveSearchIndex,
		func(idx indexpkg.Index, input string) ([]engine.Result, error) {
			return []engine.Result{
				{
					Profile: indexpkg.Profile{
						Username: "userA",
						Signals: map[string]float64{
							"consistency": 0.82,
							"ownership":   0.75,
							"depth":       0.68,
							"activity":    0.90,
						},
					},
					Score:   0.87,
					Reasons: []string{"High consistency (0.82 >= 0.70)", "Strong ownership (0.75 >= 0.60)"},
				},
			}, nil
		},
	)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/search?q=backend", nil)
	rec := httptest.NewRecorder()

	searchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload searchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if payload.Query != "backend" {
		t.Fatalf("expected query backend, got %q", payload.Query)
	}
	if payload.Mode != "dataset" {
		t.Fatalf("expected mode dataset, got %q", payload.Mode)
	}
	if payload.Total != 1 {
		t.Fatalf("expected total 1, got %d", payload.Total)
	}
	if len(payload.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(payload.Results))
	}

	got := payload.Results[0]
	if got.Username != "userA" {
		t.Fatalf("expected username userA, got %q", got.Username)
	}
	if got.Score != 0.87 {
		t.Fatalf("expected score 0.87, got %.2f", got.Score)
	}
	if got.Confidence != "high" {
		t.Fatalf("expected confidence high, got %q", got.Confidence)
	}
	if got.Signals["consistency"] != 0.82 || got.Signals["ownership"] != 0.75 || got.Signals["depth"] != 0.68 || got.Signals["activity"] != 0.90 {
		t.Fatalf("unexpected signals payload: %+v", got.Signals)
	}
	if len(got.Reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(got.Reasons))
	}
}

func TestSearchHandlerLiveModeEmptyCandidates(t *testing.T) {
	restore := overrideServerSearchDependencies(
		loadSearchDataset,
		func(query string) (indexpkg.Index, error) {
			return indexpkg.Index{}, nil
		},
		func(idx indexpkg.Index, input string) ([]engine.Result, error) {
			return []engine.Result{}, nil
		},
	)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/search?q=backend&live=true", nil)
	rec := httptest.NewRecorder()

	searchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload searchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if payload.Mode != "live" {
		t.Fatalf("expected live mode, got %q", payload.Mode)
	}
	if payload.Total != 0 || len(payload.Results) != 0 {
		t.Fatalf("expected empty result set, got total=%d len=%d", payload.Total, len(payload.Results))
	}
}

func TestSearchHandlerLiveFetchFailure(t *testing.T) {
	restore := overrideServerSearchDependencies(
		loadSearchDataset,
		func(query string) (indexpkg.Index, error) {
			return indexpkg.Index{}, errors.New("boom")
		},
		runSearchQuery,
	)
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/search?q=backend&live=true", nil)
	rec := httptest.NewRecorder()

	searchHandler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}

	var payload errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if payload.Error != "failed to fetch GitHub data" {
		t.Fatalf("unexpected error message: %q", payload.Error)
	}
}

func overrideServerSearchDependencies(
	load func(path string) (indexpkg.Index, error),
	live func(query string) (indexpkg.Index, error),
	searchFn func(idx indexpkg.Index, input string) ([]engine.Result, error),
) func() {
	originalLoad := loadSearchDataset
	originalLive := buildLiveSearchIndex
	originalSearch := runSearchQuery

	loadSearchDataset = load
	buildLiveSearchIndex = live
	runSearchQuery = searchFn

	return func() {
		loadSearchDataset = originalLoad
		buildLiveSearchIndex = originalLive
		runSearchQuery = originalSearch
	}
}
