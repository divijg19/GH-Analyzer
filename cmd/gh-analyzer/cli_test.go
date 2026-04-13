package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func TestRunCLIRootHelp(t *testing.T) {
	stdout, stderr, err := captureOutput(func() error {
		return runCLI([]string{"--help"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Usage:") || !strings.Contains(stdout, "Commands:") {
		t.Fatalf("expected root help output, got: %q", stdout)
	}
}

func TestRunQueryHelp(t *testing.T) {
	stdout, stderr, err := captureOutput(func() error {
		return runQuery([]string{"--help"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Usage:") || !strings.Contains(stderr, "Examples:") {
		t.Fatalf("expected query help output, got: %q", stderr)
	}
}

func TestQueryJSONOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runQuery([]string{"--dataset", datasetPath, "--json", "consistency>=0"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []map[string]any
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
}

func TestInspectJSONOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runInspect([]string{"--dataset", datasetPath, "--json", "alice"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profile indexpkg.Profile
	if err := json.Unmarshal([]byte(stdout), &profile); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if profile.Username != "alice" {
		t.Fatalf("expected username alice, got %q", profile.Username)
	}
}

func TestDatasetJSONOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runDataset([]string{"--dataset", datasetPath, "--json"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var indexData indexpkg.Index
	if err := json.Unmarshal([]byte(stdout), &indexData); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if len(indexData.All()) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(indexData.All()))
	}
}

func TestSearchJSONOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runSearch([]string{"--dataset", datasetPath, "--json", "backend"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []engine.Result
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if strings.Contains(stdout, "(high)") || strings.Contains(stdout, "(moderate)") || strings.Contains(stdout, "(low)") {
		t.Fatalf("did not expect confidence labels in JSON output: %q", stdout)
	}
}

func TestSearchOutputIncludesConfidenceLabel(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runSearch([]string{"--dataset", datasetPath, "backend"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "(high)") {
		t.Fatalf("expected confidence label in output, got %q", stdout)
	}
}

func TestDatasetStatsOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runDataset([]string{"stats", "--dataset", datasetPath})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Dataset: "+datasetPath) {
		t.Fatalf("expected dataset path in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "consistency:") || !strings.Contains(stdout, "avg: 0.65") || !strings.Contains(stdout, "min: 0.40") || !strings.Contains(stdout, "max: 0.90") {
		t.Fatalf("expected consistency stat block, got %q", stdout)
	}
	if !strings.Contains(stdout, "ownership:") || !strings.Contains(stdout, "min: 0.50") || !strings.Contains(stdout, "max: 0.80") {
		t.Fatalf("expected ownership stat block, got %q", stdout)
	}
	if !strings.Contains(stdout, "depth:") || !strings.Contains(stdout, "avg: 0.50") || !strings.Contains(stdout, "min: 0.30") || !strings.Contains(stdout, "max: 0.70") {
		t.Fatalf("expected depth stat block, got %q", stdout)
	}
}

func TestDatasetInfoOutput(t *testing.T) {
	datasetPath := writeTestDataset(t)

	stdout, _, err := captureOutput(func() error {
		return runDataset([]string{"info", "--dataset", datasetPath})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Dataset: "+datasetPath) {
		t.Fatalf("expected dataset path in info output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Profiles: 2") {
		t.Fatalf("expected profile count in info output, got %q", stdout)
	}
	if strings.Contains(stdout, "avg:") {
		t.Fatalf("did not expect stats in info output, got %q", stdout)
	}
}

func TestRunCLIMissingDatasetErrorMessage(t *testing.T) {
	err := runCLI([]string{"dataset", "--dataset", "does-not-exist.json"})
	if err == nil {
		t.Fatal("expected missing dataset error")
	}

	message := err.Error()
	if !strings.Contains(message, "dataset not found") {
		t.Fatalf("expected dataset-not-found error, got %q", message)
	}
	if !strings.Contains(message, "Run: gh-analyzer build <usernames>") {
		t.Fatalf("expected actionable guidance, got %q", message)
	}
}

func TestParseAnalyzeArgsJSON(t *testing.T) {
	options, showHelp, err := parseAnalyzeArgs([]string{"--json", "octocat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if showHelp {
		t.Fatal("expected showHelp false")
	}
	if options.Username != "octocat" {
		t.Fatalf("expected username octocat, got %q", options.Username)
	}
	if !options.JSON {
		t.Fatal("expected JSON option true")
	}
}

func captureOutput(fn func() error) (string, string, error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, _ := os.Pipe()
	stderrReader, stderrWriter, _ := os.Pipe()

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	err := fn()

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	stdoutBytes, _ := io.ReadAll(stdoutReader)
	stderrBytes, _ := io.ReadAll(stderrReader)

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return string(stdoutBytes), string(stderrBytes), err
}

func writeTestDataset(t *testing.T) string {
	t.Helper()

	indexData := indexpkg.Index{Profiles: []indexpkg.Profile{
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
				"consistency": 0.4,
				"ownership":   0.5,
				"depth":       0.3,
				"activity":    1.0,
			},
		},
	}}

	path := filepath.Join(t.TempDir(), "dataset.json")
	if err := storage.Save(path, indexData); err != nil {
		t.Fatalf("failed to save test dataset: %v", err)
	}

	return path
}
