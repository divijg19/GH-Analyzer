package main

import (
	"errors"
	"os"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func loadDataset(path string) (indexpkg.Index, error) {
	indexData, err := storage.Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return indexpkg.Index{}, missingDatasetError(path)
		}
		return indexpkg.Index{}, err
	}

	return indexData, nil
}
