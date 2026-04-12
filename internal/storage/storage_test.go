package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "dataset.json")

	input := index.Index{Profiles: []index.Profile{
		{Username: "alice", Signals: map[string]float64{"consistency": 0.8, "ownership": 0.7, "depth": 0.6}},
		{Username: "bob", Signals: map[string]float64{"consistency": 0.5, "ownership": 0.9, "depth": 0.4}},
	}}

	if err := Save(path, input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if !reflect.DeepEqual(input, loaded) {
		t.Fatalf("round trip mismatch: expected %+v, got %+v", input, loaded)
	}
}

func TestLoadNilProfilesDefaultsToEmptySlice(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "dataset.json")

	if err := os.WriteFile(path, []byte(`{"profiles":null}`), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Profiles == nil {
		t.Fatal("expected non-nil empty profiles slice")
	}
	if len(loaded.Profiles) != 0 {
		t.Fatalf("expected empty profiles, got %d", len(loaded.Profiles))
	}
}
