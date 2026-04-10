package ghanalyzer

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	recentWindowDays      = 90
	minDepthRepoSize      = 50
	maxReasonableRepoSize = 1000.0
	scoreScale            = 100.0
	minSignalValue        = 0.0
	maxSignalValue        = 1.0
	ownershipWeight       = 0.3
	consistencyWeight     = 0.4
	depthWeight           = 0.3

	strongScoreThreshold   = 70
	moderateScoreThreshold = 40

	highlightConsistency = "Active in last 90 days"
	highlightOwnership   = "Majority original repositories"
	highlightDepth       = "Includes non-trivial projects"
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
}

type Scores struct {
	Ownership   int `json:"ownership"`
	Consistency int `json:"consistency"`
	Depth       int `json:"depth"`
	Overall     int `json:"overall"`
}

type Report struct {
	Username   string   `json:"username"`
	Scores     Scores   `json:"scores"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

func ExtractSignals(repos []Repo) Signals {
	total := len(repos)
	if total == 0 {
		return Signals{}
	}

	cutoff := time.Now().AddDate(0, 0, -recentWindowDays)
	nonForkCount := 0
	recentCount := 0
	depthCount := 0
	depthTotalSize := 0

	for _, repo := range repos {
		if !repo.Fork {
			nonForkCount++
			if repo.Size >= minDepthRepoSize {
				depthCount++
				depthTotalSize += repo.Size
			}
		}

		if !repo.UpdatedAt.Before(cutoff) {
			recentCount++
		}
	}

	depth := 0.0
	if depthCount > 0 {
		depth = float64(depthTotalSize) / float64(depthCount)
	}
	depth = normalizeDepth(depth)

	return Signals{
		Ownership:   float64(nonForkCount) / float64(total),
		Consistency: float64(recentCount) / float64(total),
		Depth:       depth,
	}
}

func normalizeDepth(size float64) float64 {
	if size <= 0 {
		return 0
	}
	if size > maxReasonableRepoSize {
		size = maxReasonableRepoSize
	}

	return size / maxReasonableRepoSize
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

func BuildReport(username string, scores Scores) Report {
	return Report{
		Username:   username,
		Scores:     scores,
		Summary:    buildSummary(scores),
		Highlights: buildHighlights(scores),
	}
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
		return nil, fmt.Errorf("github api returned status %s", resp.Status)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}
