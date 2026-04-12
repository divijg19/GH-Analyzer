package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	minReposForFullScore         = 3
	smallSampleOverallMultiplier = 0.7
	defaultDatasetPath           = "dataset.json"
	defaultQueryLimit            = 10
	defaultQueryPreset           = "strong"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.TrimSpace(os.Args[1])
	args := os.Args[2:]

	var err error
	switch command {
	case "build":
		err = runBuild(args)
	case "query":
		err = runQuery(args)
	case "inspect":
		err = runInspect(args)
	case "dataset":
		err = runDataset(args)
	default:
		err = runAnalyze(command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gh-analyzer <username>")
	fmt.Println("  gh-analyzer build user1 user2 [--file users.txt] [--out dataset.json]")
	fmt.Println("  gh-analyzer query [--consistency 0.7] [--ownership 0.6] [--depth 0.5] [--limit 10] [--dataset dataset.json] [--preset strong] [expression]")
	fmt.Println("  gh-analyzer inspect username [--dataset dataset.json]")
	fmt.Println("  gh-analyzer dataset [--dataset dataset.json]")
	fmt.Println("  build saves to dataset.json by default")
}

func missingDatasetError(path string) error {
	if path == defaultDatasetPath {
		return fmt.Errorf("dataset %q not found. Run: gh-analyzer build <usernames>", path)
	}

	return fmt.Errorf("dataset %q not found. Run: gh-analyzer build <usernames> --out %s", path, path)
}
