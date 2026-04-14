package signals

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
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

	smallSampleRepoThreshold     = 3
	smallSamplePenaltyMultiplier = 0.7
	scoreScale                   = 100.0
	minSignalValue               = 0.0
	maxSignalValue               = 1.0
	ownershipWeight              = 0.3
	consistencyWeight            = 0.4
	depthWeight                  = 0.3

	strongScoreThreshold   = 70
	moderateScoreThreshold = 40

	highlightConsistency = "Active in last 90 days"
	highlightOwnership   = "Majority original repositories"
	highlightDepth       = "Includes non-trivial projects"
	topRepoLimit         = 3
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

type Scores struct {
	Ownership   int `json:"ownership"`
	Consistency int `json:"consistency"`
	Depth       int `json:"depth"`
	Overall     int `json:"overall"`
}

type Report struct {
	Username   string    `json:"username"`
	Scores     Scores    `json:"scores"`
	Summary    string    `json:"summary"`
	Highlights []string  `json:"highlights"`
	TopRepos   []TopRepo `json:"top_repos"`

	SignalValues    Signals `json:"-"`
	HasSignalValues bool    `json:"-"`
}

type TopRepo struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type GitHubAPIError struct {
	StatusCode int
	Message    string
}

func (e GitHubAPIError) Error() string {
	return e.Message
}

func ExtractSignals(repos []Repo) Signals {
	if len(repos) == 0 {
		return Signals{}
	}

	now := time.Now()
	cutoff := now.AddDate(0, 0, -recentWindowDays)

	nonForkCount := 0
	recentNonForkCount := 0
	deepRepoCount := 0

	totalValidRepos := 0
	nonForkValidRepos := 0

	for _, repo := range repos {
		if repo.Size > 0 {
			totalValidRepos++
			if !repo.Fork {
				nonForkValidRepos++
			}
		}

		if !repo.Fork {
			nonForkCount++

			if !repo.UpdatedAt.Before(cutoff) {
				recentNonForkCount++
			}

			if repo.Size >= minDepthRepoSize {
				deepRepoCount++
			}
		}
	}

	consistency := 0.0
	depth := 0.0
	ownership := 0.0

	if nonForkCount > 0 {
		consistencyDen := maxInt(minConsistencyDen, nonForkCount)
		depthDen := maxInt(minDepthDen, nonForkCount)

		consistency = float64(recentNonForkCount) / float64(consistencyDen)
		depth = float64(deepRepoCount) / float64(depthDen)
	}

	if totalValidRepos > 0 {
		ownership = float64(nonForkValidRepos) / float64(totalValidRepos)
	}

	activity := activityFromLatestRepo(repos, now)

	return Signals{
		Ownership:   clamp01(ownership),
		Consistency: clamp01(consistency),
		Depth:       clamp01(depth),
		Activity:    clamp01(activity),
	}
}

func ScoreSignals(signals Signals) Scores {
	ownership := signalToScore(signals.Ownership)
	consistency := signalToScore(signals.Consistency)
	depth := signalToScore(signals.Depth)

	overallFloat :=
		(float64(consistency) * consistencyWeight) +
			(float64(ownership) * ownershipWeight) +
			(float64(depth) * depthWeight)

	return Scores{
		Ownership:   ownership,
		Consistency: consistency,
		Depth:       depth,
		Overall:     int(math.Round(overallFloat)),
	}
}

func BuildReport(username string, scores Scores, repos []Repo) Report {
	signalValues := ExtractSignals(repos)
	scores = applySmallSamplePenalty(scores, len(repos))

	return Report{
		Username:        username,
		Scores:          scores,
		Summary:         buildSummary(scores),
		Highlights:      buildHighlights(scores),
		TopRepos:        extractTopRepos(repos),
		SignalValues:    signalValues,
		HasSignalValues: true,
	}
}

func applySmallSamplePenalty(scores Scores, repoCount int) Scores {
	if repoCount >= smallSampleRepoThreshold {
		return scores
	}

	baselineOverall := int(math.Round(
		(float64(scores.Consistency) * consistencyWeight) +
			(float64(scores.Ownership) * ownershipWeight) +
			(float64(scores.Depth) * depthWeight),
	))

	if scores.Overall < baselineOverall {
		return scores
	}

	scores.Overall = int(math.Round(float64(scores.Overall) * smallSamplePenaltyMultiplier))
	return scores
}

func activityFromLatestRepo(repos []Repo, now time.Time) float64 {
	if len(repos) == 0 {
		return 0
	}

	latest := repos[0].UpdatedAt
	for _, repo := range repos[1:] {
		if repo.UpdatedAt.After(latest) {
			latest = repo.UpdatedAt
		}
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

func extractTopRepos(repos []Repo) []TopRepo {
	topRepos := make([]TopRepo, 0, len(repos))

	for _, repo := range repos {
		if repo.Fork {
			continue
		}

		topRepos = append(topRepos, TopRepo{
			Name: repo.Name,
			Size: repo.Size,
		})
	}

	sort.Slice(topRepos, func(i, j int) bool {
		return topRepos[i].Size > topRepos[j].Size
	})

	if len(topRepos) > topRepoLimit {
		topRepos = topRepos[:topRepoLimit]
	}

	return topRepos
}

func buildSummary(scores Scores) string {
	consistency := scoreLevel(scores.Consistency)
	ownership := strings.ToLower(scoreLevel(scores.Ownership))
	depth := strings.ToLower(scoreLevel(scores.Depth))

	return fmt.Sprintf("%s consistency, %s ownership, %s depth", consistency, ownership, depth)
}

func buildHighlights(scores Scores) []string {
	highlights := []string{}

	if scores.Consistency > strongScoreThreshold {
		highlights = append(highlights, highlightConsistency)
	}
	if scores.Ownership > strongScoreThreshold {
		highlights = append(highlights, highlightOwnership)
	}
	if scores.Depth > strongScoreThreshold {
		highlights = append(highlights, highlightDepth)
	}

	return highlights
}

func scoreLevel(score int) string {
	if score > strongScoreThreshold {
		return "Strong"
	}
	if score >= moderateScoreThreshold {
		return "Moderate"
	}

	return "Low"
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

func FetchRepos(username string) ([]Repo, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gh-analyzer")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var githubError struct {
			Message string `json:"message"`
		}

		message := resp.Status
		if err := json.NewDecoder(resp.Body).Decode(&githubError); err == nil && strings.TrimSpace(githubError.Message) != "" {
			message = strings.TrimSpace(githubError.Message)
		}

		return nil, GitHubAPIError{
			StatusCode: resp.StatusCode,
			Message:    message,
		}
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}
