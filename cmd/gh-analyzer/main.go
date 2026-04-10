package main

import (
	"fmt"
	"os"

	ghanalyzer "github.com/divijg19/GH-Analyzer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "error: missing GitHub username")
		os.Exit(1)
	}

	username := os.Args[1]
	fmt.Printf("Analyzing %s\n", username)

	repos, err := ghanalyzer.FetchRepos(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total repos: %d\n", len(repos))

	signals := ghanalyzer.ExtractSignals(repos)
	fmt.Println("Signals:")
	fmt.Printf("Ownership: %.2f\n", signals.Ownership)
	fmt.Printf("Consistency: %.2f\n", signals.Consistency)
	fmt.Printf("Depth: %.2f\n", signals.Depth)
}
