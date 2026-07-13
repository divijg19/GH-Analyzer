package indicators

import (
	"math"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/observations"
)

const (
	SignalOwnership   = "ownership"
	SignalConsistency = "consistency"
	SignalDepth       = "depth"
	SignalActivity    = "activity"
	minSignalValue    = 0.0
	maxSignalValue    = 1.0
	scoreScale        = 100.0
)

type Signals struct {
	Ownership   float64
	Consistency float64
	Depth       float64
	Activity    float64
}

type RawScore struct {
	Ownership   int `json:"ownership"`
	Consistency int `json:"consistency"`
	Depth       int `json:"depth"`
}

func SignalsToMap(s Signals) map[string]float64 {
	return map[string]float64{
		SignalOwnership:   clamp01(s.Ownership),
		SignalConsistency: clamp01(s.Consistency),
		SignalDepth:       clamp01(s.Depth),
		SignalActivity:    clamp01(s.Activity),
	}
}

func FromMap(m map[string]float64) Signals {
	return Signals{
		Ownership:   m[SignalOwnership],
		Consistency: m[SignalConsistency],
		Depth:       m[SignalDepth],
		Activity:    m[SignalActivity],
	}
}

func ScoreSignals(signals Signals) RawScore {
	return RawScore{
		Ownership:   score(signals.Ownership),
		Consistency: score(signals.Consistency),
		Depth:       score(signals.Depth),
	}
}

func ExtractSignals(repos []observations.RepositoryVestige, referenceTime time.Time) Signals {
	return ExtractSignalsFromFacts(facts.FromRepos(repos, referenceTime), referenceTime)
}

func ExtractSignalsFromFacts(f facts.RepositoryFacts, referenceTime time.Time) Signals {
	consistency := 0.0
	depth := 0.0
	ownership := 0.0

	if f.OriginalRepos > 0 {
		consistencyDen := maxInt(10, f.OriginalRepos)
		depthDen := maxInt(5, f.OriginalRepos)
		consistency = float64(f.RecentRepos) / float64(consistencyDen)
		depth = float64(f.DeepRepos) / float64(depthDen)
	}
	if f.ValidRepos > 0 {
		ownership = float64(f.ValidOriginalRepos) / float64(f.ValidRepos)
	}
	activity := activityFromTime(f.LatestActivity, referenceTime)

	return Signals{Ownership: clamp01(ownership), Consistency: clamp01(consistency), Depth: clamp01(depth), Activity: clamp01(activity)}
}

func score(value float64) int {
	return int(math.Round(value * scoreScale))
}

func clamp01(value float64) float64 {
	if value < minSignalValue {
		return minSignalValue
	}
	if value > maxSignalValue {
		return maxSignalValue
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func activityFromTime(latest time.Time, now time.Time) float64 {
	if latest.IsZero() {
		return 0
	}
	daysSinceLastUpdate := int(now.Sub(latest).Hours() / 24)
	switch {
	case daysSinceLastUpdate <= 30:
		return 1.0
	case daysSinceLastUpdate <= 90:
		return 0.7
	case daysSinceLastUpdate <= 180:
		return 0.4
	default:
		return 0.1
	}
}
