package main

import (
	"fmt"
	"io"
	"strings"
)

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas <username> [--json]")
	fmt.Fprintln(w, "  atlas <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  analyze      Analyze a single GitHub profile")
	fmt.Fprintln(w, "  intelligence Canonical Candidate Intelligence view of a profile")
	fmt.Fprintln(w, "  build        Build a dataset from usernames")
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
	fmt.Fprintln(w, "  atlas kamikuma")
	fmt.Fprintln(w, "  atlas analyze --json kamikuma")
	fmt.Fprintln(w, "  atlas intelligence divijg19")
	fmt.Fprintln(w, "  atlas intelligence divijg19 --json")
	fmt.Fprintln(w, "  atlas build user1 user2 --out dataset.json")
	fmt.Fprintln(w, "  atlas query --preset strong --limit 5")
	fmt.Fprintln(w, "  atlas search backend")
	fmt.Fprintln(w, "  atlas inspect octocat --dataset dataset.json")
	fmt.Fprintln(w, "  atlas dataset info")
	fmt.Fprintln(w, "  atlas dataset preview")
	fmt.Fprintln(w, "  atlas dataset stats")
	fmt.Fprintln(w, "  atlas dataset --json --dataset dataset.json")
}

func printBuildHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas build [usernames...] [--file users.txt] [--out dataset.json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --file string   Path to username list file")
	fmt.Fprintln(w, "  --out string    Output dataset path (default: dataset.json)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas build user1 user2")
	fmt.Fprintln(w, "  atlas build --file users.txt --out team.json")
}

func printQueryHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas query [options] [expression]")
	fmt.Fprintln(w, "  Tip: use 'atlas search' for primary discovery workflows")
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
	fmt.Fprintln(w, "  atlas query --preset strong")
	fmt.Fprintln(w, "  atlas query consistency>=0.7 AND ownership>=0.6")
	fmt.Fprintln(w, "  atlas query --json --limit 5 depth>=0.5")
	fmt.Fprintln(w, "  atlas query -k 5 --compact consistency>=0.7")
}

func printInspectHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas inspect <username> [--dataset dataset.json] [--json] [--provenance]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json            Output JSON (always includes the full provenance graph)")
	fmt.Fprintln(w, "  --provenance      Show provenance chains in human output")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas inspect octocat")
	fmt.Fprintln(w, "  atlas inspect octocat --provenance")
	fmt.Fprintln(w, "  atlas inspect octocat --json --dataset team.json")
}

func printDatasetHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas dataset [--dataset dataset.json] [--json]")
	fmt.Fprintln(w, "  atlas dataset info [--dataset dataset.json]")
	fmt.Fprintln(w, "  atlas dataset preview [--dataset dataset.json]")
	fmt.Fprintln(w, "  atlas dataset stats [--dataset dataset.json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --json            Output JSON")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas dataset")
	fmt.Fprintln(w, "  atlas dataset info")
	fmt.Fprintln(w, "  atlas dataset preview")
	fmt.Fprintln(w, "  atlas dataset stats")
	fmt.Fprintln(w, "  atlas dataset --json --dataset team.json")
}

func printSearchHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas search [input]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --preset string   Preset: strong, consistent, deep")
	fmt.Fprintln(w, "  --limit int       Max results (default: 10)")
	fmt.Fprintln(w, "  -k int            Max results alias")
	fmt.Fprintln(w, "  --dataset string  Dataset file (default: dataset.json)")
	fmt.Fprintln(w, "  --live            Fetch temporary candidates from GitHub live search")
	fmt.Fprintln(w, "  --json            Output JSON")
	fmt.Fprintln(w, "  --compact         Compact output (no explanations)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas search backend")
	fmt.Fprintln(w, "  atlas search backend --live")
	fmt.Fprintln(w, "  atlas search --preset strong")
	fmt.Fprintln(w, "  atlas search \"consistency > 0.7 AND depth > 0.6\"")
	fmt.Fprintln(w, "  atlas search active --json")
	fmt.Fprintln(w, "  atlas search -k 5 --compact systems")
}

func printAnalyzeHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas <username> [--json]")
	fmt.Fprintln(w, "  atlas analyze [--json] <username>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --json   Output JSON only")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas octocat")
	fmt.Fprintln(w, "  atlas octocat --json")
	fmt.Fprintln(w, "  atlas analyze --json octocat")
}

func printIntelligenceHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  atlas intelligence [--json] <username>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --json   Output the CandidateIntelligence object as JSON")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Description:")
	fmt.Fprintln(w, "  Exposes the canonical Candidate Intelligence domain object for a user.")
	fmt.Fprintln(w, "  This is the reference consumer: it serializes the deterministic semantic")
	fmt.Fprintln(w, "  interpretation of the Profile without added presentation or interpretation.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  atlas intelligence divijg19")
	fmt.Fprintln(w, "  atlas intelligence divijg19 --json")
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
	case "intelligence":
		printIntelligenceHelp(w)
	case "dataset":
		printDatasetHelp(w)
	default:
		return fmt.Errorf("unknown command %q", command)
	}

	return nil
}
