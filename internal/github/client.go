package github

import (
	"net/http"
	"os"
)

const defaultUserAgent = "gh-analyzer"

func AuthHeader() (string, string) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", ""
	}
	return "Authorization", "Bearer " + token
}

func SetHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
	if key, val := AuthHeader(); key != "" {
		req.Header.Set(key, val)
	}
}
