package projection

import (
	"testing"
	"time"

	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/profile"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func TestBuildCandidateProjection_DisplayNameFallback(t *testing.T) {
	p := index.Profile{
		Username: "octocat",
	}
	proj := BuildCandidateProjection(p)

	if proj.DisplayName != "octocat" {
		t.Fatalf("expected DisplayName to fallback to Username, got %q", proj.DisplayName)
	}
}

func TestBuildCandidateProjection_MetadataMapping(t *testing.T) {
	p := index.Profile{
		Username: "octocat",
		Metadata: &profile.UserMetadata{
			Name:     "Monalisa Octocat",
			Company:  "@github",
			Location: "San Francisco, CA",
		},
	}
	proj := BuildCandidateProjection(p)

	if proj.DisplayName != "Monalisa Octocat" {
		t.Fatalf("expected DisplayName from metadata, got %q", proj.DisplayName)
	}
	if proj.Company != "@github" {
		t.Fatalf("expected Company, got %q", proj.Company)
	}
	if proj.Location != "San Francisco, CA" {
		t.Fatalf("expected Location, got %q", proj.Location)
	}
}

func TestBuildCandidateProjection_FactsPreserved(t *testing.T) {
	facts := &signals.Facts{
		TotalRepos: 100,
	}
	p := index.Profile{
		Username: "octocat",
		Facts:    facts,
	}
	proj := BuildCandidateProjection(p)

	if proj.Facts == nil {
		t.Fatal("expected Facts to be preserved")
	}
	if proj.Facts.TotalRepos != 100 {
		t.Fatalf("expected TotalRepos 100, got %d", proj.Facts.TotalRepos)
	}
	if proj.Facts != facts {
		t.Fatal("expected Facts to be shallow reference, not copy")
	}
}

func TestBuildCandidateProjection_SignalsPreserved(t *testing.T) {
	signalsMap := map[string]float64{
		"consistency": 0.82,
		"ownership":   0.75,
	}
	p := index.Profile{
		Username: "octocat",
		Signals:  signalsMap,
	}
	proj := BuildCandidateProjection(p)

	if proj.Signals == nil {
		t.Fatal("expected Signals to be preserved")
	}
	if proj.Signals["consistency"] != 0.82 {
		t.Fatalf("expected consistency 0.82, got %f", proj.Signals["consistency"])
	}
	if proj.Signals["ownership"] != 0.75 {
		t.Fatalf("expected ownership 0.75, got %f", proj.Signals["ownership"])
	}
}

func TestBuildCandidateProjection_ContributionsPreserved(t *testing.T) {
	contrib := &contributions.Summary{
		TotalContributions: 150,
	}
	p := index.Profile{
		Username:      "octocat",
		Contributions: contrib,
	}
	proj := BuildCandidateProjection(p)

	if proj.Contributions == nil {
		t.Fatal("expected Contributions to be preserved")
	}
	if proj.Contributions.TotalContributions != 150 {
		t.Fatalf("expected TotalContributions 150, got %d", proj.Contributions.TotalContributions)
	}
	if proj.Contributions != contrib {
		t.Fatal("expected Contributions to be shallow reference, not copy")
	}
}

func TestBuildCandidateProjection_NilMetadata(t *testing.T) {
	p := index.Profile{
		Username: "octocat",
		Metadata: nil,
	}
	proj := BuildCandidateProjection(p)

	if proj.DisplayName != "octocat" {
		t.Fatalf("expected DisplayName fallback to Username on nil Metadata, got %q", proj.DisplayName)
	}
}

func TestBuildCandidateProjection_NilFacts(t *testing.T) {
	p := index.Profile{
		Username: "octocat",
		Facts:    nil,
	}
	proj := BuildCandidateProjection(p)

	if proj.Facts != nil {
		t.Fatal("expected Facts to remain nil")
	}
}

func TestBuildCandidateProjection_NilContributions(t *testing.T) {
	p := index.Profile{
		Username:      "octocat",
		Contributions: nil,
	}
	proj := BuildCandidateProjection(p)

	if proj.Contributions != nil {
		t.Fatal("expected Contributions to remain nil")
	}
}

func TestBuildCandidateProjection_DeterministicOutput(t *testing.T) {
	meta := &profile.UserMetadata{Name: "Test User"}
	facts := &signals.Facts{TotalRepos: 10}
	signalsMap := map[string]float64{"consistency": 0.9}
	contrib := &contributions.Summary{TotalContributions: 50}

	p := index.Profile{
		Username:      "testuser",
		Metadata:      meta,
		Facts:         facts,
		Signals:       signalsMap,
		Contributions: contrib,
	}

	for i := 0; i < 10; i++ {
		proj := BuildCandidateProjection(p)

		if proj.Username != "testuser" {
			t.Fatalf("iteration %d: unexpected Username", i)
		}
		if proj.DisplayName != "Test User" {
			t.Fatalf("iteration %d: unexpected DisplayName", i)
		}
	}
}

func TestBuildCandidateProjection_ShallowReference(t *testing.T) {
	originalSignals := map[string]float64{"consistency": 0.8}
	p := index.Profile{
		Username: "octocat",
		Signals:  originalSignals,
	}

	proj := BuildCandidateProjection(p)

	// Verify shallow reference: modifying projection Signals affects original
	proj.Signals["consistency"] = 0.9

	if p.Signals["consistency"] != 0.9 {
		t.Fatal("expected Signals to be shallow reference")
	}
}

func TestBuildCandidateProjection_FullIntegration(t *testing.T) {
	now := time.Now().UTC()
	p := index.Profile{
		Username: "fulluser",
		Metadata: &profile.UserMetadata{
			Name:      "Full Name",
			Company:   "Acme Corp",
			Location:  "New York",
			CreatedAt: now,
		},
		Facts: &signals.Facts{
			TotalRepos: 50,
		},
		Signals: map[string]float64{
			"consistency": 0.7,
			"ownership":   0.6,
		},
		Contributions: &contributions.Summary{
			TotalPullRequests: 200,
		},
	}

	proj := BuildCandidateProjection(p)

	if proj.Username != "fulluser" {
		t.Fatalf("unexpected Username: %s", proj.Username)
	}
	if proj.DisplayName != "Full Name" {
		t.Fatalf("unexpected DisplayName: %s", proj.DisplayName)
	}
	if proj.Company != "Acme Corp" {
		t.Fatalf("unexpected Company: %s", proj.Company)
	}
	if proj.Location != "New York" {
		t.Fatalf("unexpected Location: %s", proj.Location)
	}
	if proj.Facts == nil || proj.Facts.TotalRepos != 50 {
		t.Fatalf("unexpected Facts")
	}
	if proj.Signals == nil || proj.Signals["consistency"] != 0.7 {
		t.Fatalf("unexpected Signals")
	}
	if proj.Contributions == nil || proj.Contributions.TotalPullRequests != 200 {
		t.Fatalf("unexpected Contributions")
	}
}
