package signals

import (
	"math"
	"time"
)

const (
	recentWindowDays  = 90
	minDepthRepoSize  = 50
	minConsistencyDen = 10
	minDepthDen       = 5

	activityHighDays     = 30
	activityModerateDays = 90
	activityLowDays      = 180

	activityHighValue     = 1.0
	activityModerateValue = 0.7
	activityLowValue      = 0.4
	activityStaleValue    = 0.1

	scoreScale        = 100.0
	minSignalValue    = 0.0
	maxSignalValue    = 1.0
	ownershipWeight   = 0.3
	consistencyWeight = 0.4
	depthWeight       = 0.3
)

type Repo struct {
	Name      string    `json:"name"`
	Fork      bool      `json:"fork"`
	Size      int       `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

type GitHubAPIError struct {
	StatusCode int
	Message    string
}

func (e GitHubAPIError) Error() string {
	return e.Message
}

func ExtractSignals(repos []Repo) Signals {
	return ExtractSignalsFromFacts(FromRepos(repos))
}

func ExtractSignalsFromFacts(f Facts) Signals {
	consistency := 0.0
	depth := 0.0
	ownership := 0.0

	if f.OriginalRepos > 0 {
		consistencyDen := maxInt(minConsistencyDen, f.OriginalRepos)
		depthDen := maxInt(minDepthDen, f.OriginalRepos)
		consistency = float64(f.RecentRepos) / float64(consistencyDen)
		depth = float64(f.DeepRepos) / float64(depthDen)
	}

	if f.ValidRepos > 0 {
		ownership = float64(f.ValidOriginalRepos) / float64(f.ValidRepos)
	}

	activity := activityFromTime(f.LatestActivity, time.Now())

	return Signals{
		Ownership:   clamp01(ownership),
		Consistency: clamp01(consistency),
		Depth:       clamp01(depth),
		Activity:    clamp01(activity),
	}
}

func ScoreSignals(signals Signals) RawScore {
	ownership := signalToScore(signals.Ownership)
	consistency := signalToScore(signals.Consistency)
	depth := signalToScore(signals.Depth)

	return RawScore{
		Ownership:   ownership,
		Consistency: consistency,
		Depth:       depth,
	}
}

func activityFromTime(latest time.Time, now time.Time) float64 {
	if latest.IsZero() {
		return 0
	}

	daysSinceLastUpdate := int(now.Sub(latest).Hours() / 24)

	switch {
	case daysSinceLastUpdate <= activityHighDays:
		return activityHighValue
	case daysSinceLastUpdate <= activityModerateDays:
		return activityModerateValue
	case daysSinceLastUpdate <= activityLowDays:
		return activityLowValue
	default:
		return activityStaleValue
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func signalToScore(value float64) int {
	if value < minSignalValue {
		value = minSignalValue
	}
	if value > maxSignalValue {
		value = maxSignalValue
	}

	scaled := value * scoreScale
	return int(math.Round(scaled))
}
