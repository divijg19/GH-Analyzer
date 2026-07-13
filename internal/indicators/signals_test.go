package indicators

import (
	"reflect"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
)

// TestSignalsClampedToUnitInterval verifies the documented invariant: signal
// values are always clamped to [0, 1].
func TestSignalsClampedToUnitInterval(t *testing.T) {
	ref := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// Facts engineered so naive formulas would exceed [0,1] if unclamped.
	f := facts.RepositoryFacts{
		OriginalRepos:      1,
		RecentRepos:        1, // consistency = 1/10 => 0.1 (well within range)
		ValidRepos:         1,
		ValidOriginalRepos: 1,   // ownership = 1/1 => 1.0 (upper bound)
		DeepRepos:          1,   // depth = 1/5 => 0.2
		LatestActivity:     ref, // activity at reference => 1.0 (upper bound)
	}

	s := ExtractSignalsFromFacts(f, ref)
	if s.Ownership < 0 || s.Ownership > 1 {
		t.Fatalf("Ownership out of [0,1]: %v", s.Ownership)
	}
	if s.Depth < 0 || s.Depth > 1 {
		t.Fatalf("Depth out of [0,1]: %v", s.Depth)
	}
	if s.Activity < 0 || s.Activity > 1 {
		t.Fatalf("Activity out of [0,1]: %v", s.Activity)
	}
	if s.Ownership != 1.0 {
		t.Fatalf("expected saturated ownership 1.0, got %v", s.Ownership)
	}
	if s.Activity != 1.0 {
		t.Fatalf("expected saturated activity 1.0, got %v", s.Activity)
	}
}

// TestExtractSignalsFromFactsDeterministic verifies the documented invariant:
// identical facts plus identical referenceTime always yield identical signals.
func TestExtractSignalsFromFactsDeterministic(t *testing.T) {
	ref := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	f := facts.RepositoryFacts{
		OriginalRepos:      10,
		RecentRepos:        5,
		ValidRepos:         8,
		ValidOriginalRepos: 6,
		DeepRepos:          4,
		LatestActivity:     ref.AddDate(0, 0, -10),
	}

	first := ExtractSignalsFromFacts(f, ref)
	for i := 0; i < 1000; i++ {
		if got := ExtractSignalsFromFacts(f, ref); !reflect.DeepEqual(first, got) {
			t.Fatalf("ExtractSignalsFromFacts not deterministic on iteration %d", i)
		}
	}
}

// TestScoreSignalsScalesTo100 verifies RawScore maps the unit interval to
// [0, 100].
func TestScoreSignalsScalesTo100(t *testing.T) {
	s := Signals{Ownership: 1.0, Consistency: 1.0, Depth: 1.0, Activity: 1.0}
	rs := ScoreSignals(s)
	if rs.Ownership != 100 || rs.Consistency != 100 || rs.Depth != 100 {
		t.Fatalf("expected all scores 100, got %+v", rs)
	}

	s2 := Signals{Ownership: 0.5, Consistency: 0.5, Depth: 0.5, Activity: 0.5}
	rs2 := ScoreSignals(s2)
	if rs2.Ownership != 50 || rs2.Consistency != 50 || rs2.Depth != 50 {
		t.Fatalf("expected all scores 50, got %+v", rs2)
	}
}

// TestSignalsToMapFromMapRoundTrip verifies the canonical map conversion is
// lossless for the four known signals.
func TestSignalsToMapFromMapRoundTrip(t *testing.T) {
	s := Signals{Ownership: 0.3, Consistency: 0.7, Depth: 0.9, Activity: 0.4}
	m := SignalsToMap(s)
	if m[SignalOwnership] != 0.3 || m[SignalConsistency] != 0.7 ||
		m[SignalDepth] != 0.9 || m[SignalActivity] != 0.4 {
		t.Fatalf("SignalsToMap dropped/clamped values: %v", m)
	}
	back := FromMap(m)
	if !reflect.DeepEqual(s, back) {
		t.Fatalf("round trip mismatch: %+v != %+v", s, back)
	}
}
