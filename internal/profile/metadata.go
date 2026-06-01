package profile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/divijg19/GH-Analyzer/internal/github"
)

type UserMetadata struct {
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	Location  string    `json:"location"`
	Company   string    `json:"company"`
	Followers int       `json:"followers"`
	Following int       `json:"following"`
	CreatedAt time.Time `json:"created_at"`
}

var githubUserURL = func(username string) string {
	return fmt.Sprintf("https://api.github.com/users/%s", username)
}

func FetchUserMetadata(username string) (*UserMetadata, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	url := githubUserURL(username)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	github.SetHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var ghErr struct {
			Message string `json:"message"`
		}

		message := resp.Status
		if err := json.NewDecoder(resp.Body).Decode(&ghErr); err == nil && ghErr.Message != "" {
			message = ghErr.Message
		}

		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, message)
	}

	var raw struct {
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		Location  string `json:"location"`
		Company   string `json:"company"`
		Followers int    `json:"followers"`
		Following int    `json:"following"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode user metadata: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		createdAt = time.Time{}
	}

	return &UserMetadata{
		Name:      raw.Name,
		Bio:       raw.Bio,
		Location:  raw.Location,
		Company:   raw.Company,
		Followers: raw.Followers,
		Following: raw.Following,
		CreatedAt: createdAt,
	}, nil
}
