package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/signals"
)

const (
	serverAddr                   = ":8080"
	minReposForFullScore         = 3
	smallSampleOverallMultiplier = 0.7
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: message}); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "missing username")
		return
	}

	repos, err := signals.FetchRepos(username)
	if err != nil {
		var githubAPIError signals.GitHubAPIError
		if errors.As(err, &githubAPIError) {
			switch githubAPIError.StatusCode {
			case http.StatusNotFound:
				writeJSONError(w, http.StatusNotFound, "GitHub user not found")
			case http.StatusForbidden:
				if strings.Contains(strings.ToLower(githubAPIError.Message), "rate limit") {
					writeJSONError(w, http.StatusTooManyRequests, "GitHub API rate limit exceeded. Please try again later.")
					break
				}
				writeJSONError(w, http.StatusBadGateway, "GitHub API access denied")
			default:
				writeJSONError(w, http.StatusBadGateway, "GitHub API request failed")
			}
			return
		}

		writeJSONError(w, http.StatusBadGateway, "unable to reach GitHub API")
		return
	}

	signalValues := signals.ExtractSignals(repos)
	scores := signals.ScoreSignals(signalValues)
	if len(repos) < minReposForFullScore {
		scores.Overall = int(math.Round(float64(scores.Overall) * smallSampleOverallMultiplier))
	}

	report := signals.BuildReport(username, scores, repos)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	fmt.Println("server running on :8080")
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		panic(err)
	}
}
