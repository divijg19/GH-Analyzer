package projection

import (
	"testing"
	"time"

	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/profile"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func TestBuildInspectProjection_DeterministicOutput(t *testing.T) {
	now := time.Now().UTC()
	p := index.Profile{
		Username: "testuser",
		Metadata: &profile.UserMetadata{
			Name:     "Test User",
			Company:  "Acme",
			Location: "NYC",
		},
		Facts: &signals.Facts{
			TotalRepos:         100,
			OriginalRepos:      80,
			ForkRepos:          20,
			RecentRepos:        50,
			DeepRepos:          30,
			ValidRepos:         90,
			ValidOriginalRepos: 70,
			LargestRepoSize:    5000,
			LatestActivity:     now,
		},
		Signals: map[string]float64{
			"ownership":   0.78,
			"consistency": 0.56,
			"depth":       0.33,
			"activity":    0.45,
		},
		Contributions: &contributions.Summary{
			TotalContributions: 200,
			TotalPullRequests:  150,
			IssuesOpened:       30,
		},
	}

	for i := 0; i < 10; i++ {
		proj := BuildInspectProjection(p)

		if proj.Username != "testuser" {
			t.Fatalf("iteration %d: unexpected Username", i)
		}
		if proj.Metadata == nil {
			t.Fatalf("iteration %d: expected Metadata", i)
		}
		if proj.Facts == nil {
			t.Fatalf("iteration %d: expected Facts", i)
		}
		if proj.Contributions == nil {
			t.Fatalf("iteration %d: expected Contributions", i)
		}
		if len(proj.Evidence) == 0 {
			t.Fatalf("iteration %d: expected Evidence", i)
		}
	}
}

func TestBuildInspectProjection_EvidenceGeneration(t *testing.T) {
	now := time.Now().UTC()
	p := index.Profile{
		Username: "testuser",
		Facts: &signals.Facts{
			TotalRepos:         100,
			OriginalRepos:      80,
			RecentRepos:        50,
			DeepRepos:          30,
			ValidRepos:         90,
			ValidOriginalRepos: 70,
			LargestRepoSize:    5000,
			LatestActivity:     now,
		},
		Signals: map[string]float64{
			"ownership":   0.78,
			"consistency": 0.56,
			"depth":       0.33,
			"activity":    0.45,
		},
	}

	proj := BuildInspectProjection(p)

	if len(proj.Evidence) != 4 {
		t.Fatalf("expected 4 evidence groups, got %d", len(proj.Evidence))
	}

	signalNames := make(map[string]bool)
	for _, group := range proj.Evidence {
		signalNames[group.Signal] = true
		if len(group.Items) == 0 {
			t.Fatalf("expected evidence items for signal %s", group.Signal)
		}
	}

	for _, expected := range []string{"ownership", "consistency", "depth", "activity"} {
		if !signalNames[expected] {
			t.Fatalf("missing evidence group for signal %s", expected)
		}
	}
}

func TestBuildInspectProjection_NilHandling(t *testing.T) {
	p := index.Profile{
		Username: "testuser",
	}

	proj := BuildInspectProjection(p)

	if proj.Username != "testuser" {
		t.Fatalf("unexpected Username: %s", proj.Username)
	}
	if proj.Metadata != nil {
		t.Fatal("expected nil Metadata")
	}
	if proj.Facts != nil {
		t.Fatal("expected nil Facts")
	}
	if proj.Contributions != nil {
		t.Fatal("expected nil Contributions")
	}
	if len(proj.Evidence) != 0 {
		t.Fatalf("expected empty Evidence, got %d groups", len(proj.Evidence))
	}
}

func TestBuildInspectProjection_NilFactsNoEvidence(t *testing.T) {
	p := index.Profile{
		Username: "testuser",
		Signals: map[string]float64{
			"ownership": 0.5,
		},
	}

	proj := BuildInspectProjection(p)

	if len(proj.Evidence) != 0 {
		t.Fatalf("expected empty Evidence when Facts is nil, got %d groups", len(proj.Evidence))
	}
}

func TestBuildInspectProjection_MetadataMapping(t *testing.T) {
	p := index.Profile{
		Username: "testuser",
		Metadata: &profile.UserMetadata{
			Name:      "Test User",
			Company:   "Acme",
			Location:  "NYC",
			CreatedAt: time.Now().UTC(),
		},
	}

	proj := BuildInspectProjection(p)

	if proj.Metadata == nil {
		t.Fatal("expected Metadata")
	}
	if proj.Metadata.Name != "Test User" {
		t.Fatalf("expected Name 'Test User', got %q", proj.Metadata.Name)
	}
	if proj.Metadata.Company != "Acme" {
		t.Fatalf("expected Company 'Acme', got %q", proj.Metadata.Company)
	}
	if proj.Metadata.Location != "NYC" {
		t.Fatalf("expected Location 'NYC', got %q", proj.Metadata.Location)
	}
}

func TestBuildInspectProjection_FactsMapping(t *testing.T) {
	facts := &signals.Facts{
		TotalRepos:    100,
		OriginalRepos: 80,
		ForkRepos:     20,
	}
	p := index.Profile{
		Username: "testuser",
		Facts:    facts,
	}

	proj := BuildInspectProjection(p)

	if proj.Facts == nil {
		t.Fatal("expected Facts")
	}
	if proj.Facts.TotalRepos != 100 {
		t.Fatalf("expected TotalRepos 100, got %d", proj.Facts.TotalRepos)
	}
	if proj.Facts.OriginalRepos != 80 {
		t.Fatalf("expected OriginalRepos 80, got %d", proj.Facts.OriginalRepos)
	}
	if proj.Facts.ForkRepos != 20 {
		t.Fatalf("expected ForkRepos 20, got %d", proj.Facts.ForkRepos)
	}
	if proj.Facts != facts {
		t.Fatal("expected Facts to be shallow reference")
	}
}

func TestBuildInspectProjection_ContributionsMapping(t *testing.T) {
	contrib := &contributions.Summary{
		TotalContributions: 200,
		TotalPullRequests:  150,
		IssuesOpened:       30,
	}
	p := index.Profile{
		Username:      "testuser",
		Contributions: contrib,
	}

	proj := BuildInspectProjection(p)

	if proj.Contributions == nil {
		t.Fatal("expected Contributions")
	}
	if proj.Contributions.TotalContributions != 200 {
		t.Fatalf("expected TotalContributions 200, got %d", proj.Contributions.TotalContributions)
	}
	if proj.Contributions.TotalPullRequests != 150 {
		t.Fatalf("expected TotalPullRequests 150, got %d", proj.Contributions.TotalPullRequests)
	}
	if proj.Contributions.IssuesOpened != 30 {
		t.Fatalf("expected IssuesOpened 30, got %d", proj.Contributions.IssuesOpened)
	}
	if proj.Contributions != contrib {
		t.Fatal("expected Contributions to be shallow reference")
	}
}

func TestBuildInspectProjection_SignalsPreserved(t *testing.T) {
	signalsMap := map[string]float64{
		"ownership":   0.78,
		"consistency": 0.56,
		"depth":       0.33,
		"activity":    0.45,
	}
	p := index.Profile{
		Username: "testuser",
		Signals:  signalsMap,
	}

	proj := BuildInspectProjection(p)

	if proj.Signals == nil {
		t.Fatal("expected Signals")
	}
	if proj.Signals["ownership"] != 0.78 {
		t.Fatalf("expected ownership 0.78, got %f", proj.Signals["ownership"])
	}
	if proj.Signals["consistency"] != 0.56 {
		t.Fatalf("expected consistency 0.56, got %f", proj.Signals["consistency"])
	}
	if proj.Signals["depth"] != 0.33 {
		t.Fatalf("expected depth 0.33, got %f", proj.Signals["depth"])
	}
	if proj.Signals["activity"] != 0.45 {
		t.Fatalf("expected activity 0.45, got %f", proj.Signals["activity"])
	}
}

func TestBuildInspectProjection_FullIntegration(t *testing.T) {
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
			TotalRepos:         50,
			OriginalRepos:      40,
			RecentRepos:        30,
			DeepRepos:          20,
			ValidRepos:         45,
			ValidOriginalRepos: 35,
			LargestRepoSize:    10000,
			LatestActivity:     now,
		},
		Signals: map[string]float64{
			"ownership":   0.7,
			"consistency": 0.6,
			"depth":       0.5,
			"activity":    0.4,
		},
		Contributions: &contributions.Summary{
			TotalContributions: 100,
			TotalPullRequests:  80,
			IssuesOpened:       20,
		},
	}

	proj := BuildInspectProjection(p)

	if proj.Username != "fulluser" {
		t.Fatalf("unexpected Username: %s", proj.Username)
	}
	if proj.Metadata == nil || proj.Metadata.Name != "Full Name" {
		t.Fatalf("unexpected Metadata")
	}
	if proj.Facts == nil || proj.Facts.TotalRepos != 50 {
		t.Fatalf("unexpected Facts")
	}
	if proj.Signals == nil || proj.Signals["ownership"] != 0.7 {
		t.Fatalf("unexpected Signals")
	}
	if proj.Contributions == nil || proj.Contributions.TotalContributions != 100 {
		t.Fatalf("unexpected Contributions")
	}
	if len(proj.Evidence) != 4 {
		t.Fatalf("expected 4 evidence groups, got %d", len(proj.Evidence))
	}
}
