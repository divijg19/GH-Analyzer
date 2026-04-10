package ghanalyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func ExtractSignals(repos []Repo) Signals {
	total := len(repos)
	if total == 0 {
		return Signals{}
	}

	cutoff := time.Now().AddDate(0, 0, -90)
	nonForkCount := 0
	recentCount := 0
	depthCount := 0
	depthTotalSize := 0

	for _, repo := range repos {
		if !repo.Fork {
			nonForkCount++
			if repo.Size >= 50 {
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

	return Signals{
		Ownership:   float64(nonForkCount) / float64(total),
		Consistency: float64(recentCount) / float64(total),
		Depth:       depth,
	}
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