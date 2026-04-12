package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	filePath := fs.String("file", "", "path to username list file")
	outPath := fs.String("out", "", "output dataset path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	usernames, err := collectUsernames(fs.Args(), *filePath)
	if err != nil {
		return err
	}

	indexData, err := indexpkg.Build(usernames)
	if err != nil {
		return err
	}

	fmt.Printf("Built index with %d profiles\n", len(indexData.All()))
	printAverageSignals(indexData)

	resolvedOutPath := strings.TrimSpace(*outPath)
	if resolvedOutPath == "" {
		resolvedOutPath = defaultDatasetPath
	}

	if err := storage.Save(resolvedOutPath, indexData); err != nil {
		return err
	}
	fmt.Printf("Saved to %s\n", resolvedOutPath)

	return nil
}

func collectUsernames(positional []string, filePath string) ([]string, error) {
	usernames := append([]string{}, positional...)

	if strings.TrimSpace(filePath) != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			name := strings.TrimSpace(scanner.Text())
			if name == "" {
				continue
			}
			usernames = append(usernames, name)
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	clean := make([]string, 0, len(usernames))
	for _, name := range usernames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}

	if len(clean) == 0 {
		return nil, fmt.Errorf("no usernames provided")
	}

	return clean, nil
}
