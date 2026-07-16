package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/acquisition"
	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/evaluation"
	indexpkg "github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/projection"
	searchpkg "github.com/divijg19/Atlas/internal/search"
	"github.com/divijg19/Atlas/internal/storage"
)

const (
	serverAddr         = ":8080"
	defaultDatasetPath = "dataset.json"
)

var (
	loadSearchDataset = func(path string) (indexpkg.Index, error) {
		return storage.Load(path)
	}
	buildLiveSearchIndex = buildLiveIndexForServer
	runSearchQuery       = func(idx indexpkg.Index, input string) ([]engine.Result, error) {
		return searchpkg.Search(idx, input, searchpkg.Options{Limit: 0})
	}
	// buildAnalyzeProfile assembles a candidate Profile through the canonical
	// index layer rather than reaching into acquisition directly, keeping the
	// server a pure presentation surface.
	buildAnalyzeProfile = func(ctx context.Context, username string) (indexpkg.Profile, error) {
		return indexpkg.BuildProfile(ctx, acquisition.NewClient(), username, time.Now())
	}
)

type errorResponse struct {
	Error string `json:"error"`
}

type searchResponse struct {
	Query   string                        `json:"query"`
	Mode    string                        `json:"mode"`
	Total   int                           `json:"total"`
	Results []projection.SearchProjection `json:"results"`
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

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Assemble the candidate Profile through the canonical index layer. The
	// server is a presentation surface: it never acquires or normalizes
	// directly (see docs/ARCHITECTURE.md presentation boundary).
	profile, err := buildAnalyzeProfile(ctx, username)
	if err != nil {
		// Acquisition errors surface wrapped as acquisition.APIError; map the
		// status code to an appropriate HTTP response without leaking transport.
		var githubAPIError acquisition.APIError
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

	proj := projection.BuildAnalyzeProjection(profile)

	writeJSONResponse(w, http.StatusOK, proj)
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
		liveIndex, err := buildLiveSearchIndex(r.Context(), query)
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

	projections := make([]projection.SearchProjection, len(results))
	for i, result := range results {
		projections[i] = projection.BuildSearchProjection(
			result.Profile,
			result.Score,
			evaluation.ClassifyConfidence(result.Score),
			result.Reasons,
		)
	}

	response := searchResponse{
		Query:   query,
		Mode:    mode,
		Total:   len(projections),
		Results: projections,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/search", searchHandler)
	fmt.Println("server running on :8080")
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		panic(err)
	}
}
