package ghanalyzer

import (
	"encoding/json"
	"os"
)

func Save(path string, idx Index) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load(path string) (Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Index{}, err
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return Index{}, err
	}

	if idx.Profiles == nil {
		idx.Profiles = []Profile{}
	}

	return idx, nil
}
