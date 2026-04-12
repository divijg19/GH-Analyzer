package storage

import (
	"encoding/json"
	"os"

	idx "github.com/divijg19/GH-Analyzer/internal/index"
)

func Save(path string, index idx.Index) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load(path string) (idx.Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return idx.Index{}, err
	}

	var index idx.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return idx.Index{}, err
	}

	if index.Profiles == nil {
		index.Profiles = []idx.Profile{}
	}

	return index, nil
}
