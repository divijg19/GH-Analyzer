package main

import (
	"fmt"
	"io"
	"strings"
)

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer <username> [--json]")
	fmt.Fprintln(w, "  gh-analyzer <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  analyze  Analyze a single GitHub profile")
	fmt.Fprintln(w, "  build    Build a dataset from usernames")
	fmt.Fprintln(w, "  query    Advanced signal-threshold query")
	fmt.Fprintln(w, "  search   Discover developers by intent or expression")
	fmt.Fprintln(w, "  inspect  Inspect a single profile in a dataset")
	fmt.Fprintln(w, "  dataset  Show dataset summary")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Global Options:")
	fmt.Fprintln(w, "  --help      Show this help")
	fmt.Fprintln(w, "  --version   Show version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer kamikuma")
	fmt.Fprintln(w, "  gh-analyzer analyze --json kamikuma")
	fmt.Fprintln(w, "  gh-analyzer build user1 user2 --out dataset.json")
	fmt.Fprintln(w, "  gh-analyzer query --preset strong --limit 5")
	fmt.Fprintln(w, "  gh-analyzer search backend")
	fmt.Fprintln(w, "  gh-analyzer inspect octocat --dataset dataset.json")
	fmt.Fprintln(w, "  gh-analyzer dataset info")
	fmt.Fprintln(w, "  gh-analyzer dataset preview")
	fmt.Fprintln(w, "  gh-analyzer dataset stats")
	fmt.Fprintln(w, "  gh-analyzer dataset --json --dataset dataset.json")
}

func printBuildHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer build [usernames...] [--file users.txt] [--out dataset.json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --file string   Path to username list file")
	fmt.Fprintln(w, "  --out string    Output dataset path (default: dataset.json)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer build user1 user2")
	fmt.Fprintln(w, "  gh-analyzer build --file users.txt --out team.json")
}

func printQueryHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer query [options] [expression]")
	fmt.Fprintln(w, "  Tip: use 'gh-analyzer search' for primary discovery workflows")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --consistency float  Consistency threshold")
	fmt.Fprintln(w, "  --ownership float    Ownership threshold")
	fmt.Fprintln(w, "  --depth float        Depth threshold")
	fmt.Fprintln(w, "  --preset string      Preset: strong, consistent, deep")
	fmt.Fprintln(w, "  --limit int          Max results (default: 10)")
	fmt.Fprintln(w, "  -k int               Max results alias")
	fmt.Fprintln(w, "  --dataset string     Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json               Output JSON")
	fmt.Fprintln(w, "  --compact            Compact output (no explanations)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer query --preset strong")
	fmt.Fprintln(w, "  gh-analyzer query consistency>=0.7 AND ownership>=0.6")
	fmt.Fprintln(w, "  gh-analyzer query --json --limit 5 depth>=0.5")
	fmt.Fprintln(w, "  gh-analyzer query -k 5 --compact consistency>=0.7")
}

func printInspectHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer inspect <username> [--dataset dataset.json] [--json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json            Output JSON")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer inspect octocat")
	fmt.Fprintln(w, "  gh-analyzer inspect octocat --json --dataset team.json")
}

func printDatasetHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer dataset [--dataset dataset.json] [--json]")
	fmt.Fprintln(w, "  gh-analyzer dataset info [--dataset dataset.json]")
	fmt.Fprintln(w, "  gh-analyzer dataset preview [--dataset dataset.json]")
	fmt.Fprintln(w, "  gh-analyzer dataset stats [--dataset dataset.json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json            Output JSON")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer dataset")
	fmt.Fprintln(w, "  gh-analyzer dataset info")
	fmt.Fprintln(w, "  gh-analyzer dataset preview")
	fmt.Fprintln(w, "  gh-analyzer dataset stats")
	fmt.Fprintln(w, "  gh-analyzer dataset --json --dataset team.json")
}

func printSearchHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer search [input]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --preset string   Preset: strong, consistent, deep")
	fmt.Fprintln(w, "  --limit int       Max results (default: 10)")
	fmt.Fprintln(w, "  -k int            Max results alias")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json            Output JSON")
	fmt.Fprintln(w, "  --compact         Compact output (no explanations)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer search backend")
	fmt.Fprintln(w, "  gh-analyzer search --preset strong")
	fmt.Fprintln(w, "  gh-analyzer search \"consistency > 0.7 AND depth > 0.6\"")
	fmt.Fprintln(w, "  gh-analyzer search active --json")
	fmt.Fprintln(w, "  gh-analyzer search -k 5 --compact systems")
}

func printAnalyzeHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gh-analyzer <username> [--json]")
	fmt.Fprintln(w, "  gh-analyzer analyze [--json] <username>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --json   Output JSON only")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gh-analyzer octocat")
	fmt.Fprintln(w, "  gh-analyzer octocat --json")
	fmt.Fprintln(w, "  gh-analyzer analyze --json octocat")
}

func printCommandHelp(command string, w io.Writer) error {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "analyze":
		printAnalyzeHelp(w)
	case "build":
		printBuildHelp(w)
	case "query":
		printQueryHelp(w)
	case "search":
		printSearchHelp(w)
	case "inspect":
		printInspectHelp(w)
	case "dataset":
		printDatasetHelp(w)
	default:
		return fmt.Errorf("unknown command %q", command)
	}

	return nil
}
