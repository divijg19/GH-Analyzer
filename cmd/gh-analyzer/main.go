package main

import (
	"errors"
	"flag"
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
	cliVersion                   = "v0.6.3"
)

func main() {
	err := runCLI(os.Args[1:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(args []string) error {
	if len(args) == 0 {
		printRootHelp(os.Stdout)
		return fmt.Errorf("missing command or username")
	}

	first := strings.TrimSpace(args[0])
	rest := args[1:]

	switch first {
	case "--help", "-h", "help":
		if len(rest) > 0 {
			return printCommandHelp(rest[0], os.Stdout)
		}
		printRootHelp(os.Stdout)
		return nil
	case "--version", "version":
		fmt.Println(cliVersion)
		return nil
	case "build":
		return runBuild(rest)
	case "query":
		return runQuery(rest)
	case "inspect":
		return runInspect(rest)
	case "dataset":
		return runDataset(rest)
	case "search":
		return runSearch(rest)
	case "analyze":
		return runAnalyze(rest)
	default:
		// Backward-compatible mode: treat first positional token as username.
		if strings.HasPrefix(first, "-") {
			return fmt.Errorf("unknown global option %q", first)
		}
		return runAnalyze(args)
	}
}

func missingDatasetError(path string) error {
	if path == defaultDatasetPath {
		return fmt.Errorf("dataset not found\nRun: gh-analyzer build <usernames>")
	}

	return fmt.Errorf("dataset not found: %s\nRun: gh-analyzer build <usernames> --out %s", path, path)
}

func parseFlagsOrHelp(fs *flag.FlagSet, args []string) (bool, error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}
