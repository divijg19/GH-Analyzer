package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	searchpkg "github.com/divijg19/GH-Analyzer/internal/search"
	"github.com/divijg19/GH-Analyzer/internal/signals"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

const (
	serverAddr                   = ":8080"
	defaultDatasetPath           = "dataset.json"
	minReposForFullScore         = 3
	smallSampleOverallMultiplier = 0.7
)

var (
	loadSearchDataset = func(path string) (indexpkg.Index, error) {
		return storage.Load(path)
	}
	buildLiveSearchIndex = buildLiveIndexForServer
	runSearchQuery       = func(idx indexpkg.Index, input string) ([]engine.Result, error) {
		return searchpkg.Search(idx, input, searchpkg.Options{Limit: 0})
	}
)

type errorResponse struct {
	Error string `json:"error"`
}

type searchResponse struct {
	Query   string               `json:"query"`
	Mode    string               `json:"mode"`
	Total   int                  `json:"total"`
	Results []searchResultRecord `json:"results"`
}

type searchResultRecord struct {
	Username   string             `json:"username"`
	Score      float64            `json:"score"`
	Confidence string             `json:"confidence"`
	Signals    map[string]float64 `json:"signals"`
	Reasons    []string           `json:"reasons"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: message}); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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

	writeJSONResponse(w, http.StatusOK, report)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSONError(w, http.StatusBadRequest, "missing query")
		return
	}

	liveMode := false
	liveParam := strings.TrimSpace(r.URL.Query().Get("live"))
	if liveParam != "" {
		parsedLive, err := strconv.ParseBool(liveParam)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid live flag")
			return
		}
		liveMode = parsedLive
	}

	mode := "dataset"
	var idx indexpkg.Index

	if liveMode {
		mode = "live"
		liveIndex, err := buildLiveSearchIndex(query)
		if err != nil {
			writeJSONError(w, http.StatusBadGateway, "failed to fetch GitHub data")
			return
		}
		idx = liveIndex
	} else {
		dataset, err := loadSearchDataset(defaultDatasetPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeJSONError(w, http.StatusNotFound, "dataset not found")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "failed to load dataset")
			return
		}
		idx = dataset
	}

	results, err := runSearchQuery(idx, query)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	responseResults := make([]searchResultRecord, 0, len(results))
	for _, result := range results {
		record := searchResultRecord{
			Username:   result.Profile.Username,
			Score:      result.Score,
			Confidence: confidenceFromScore(result.Score),
			Signals:    cloneSignals(result.Profile.Signals),
			Reasons:    append([]string{}, result.Reasons...),
		}
		responseResults = append(responseResults, record)
	}

	response := searchResponse{
		Query:   query,
		Mode:    mode,
		Total:   len(responseResults),
		Results: responseResults,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func cloneSignals(in map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for key, value := range in {
		out[key] = value
	}

	return out
}

func confidenceFromScore(score float64) string {
	switch {
	case score > 0.75:
		return "high"
	case score > 0.50:
		return "moderate"
	default:
		return "low"
	}
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/search", searchHandler)
	fmt.Println("server running on :8080")
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		panic(err)
	}
}
