package projection

import (
	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

// CandidateProjection is a minimal view for candidate listings.
// It represents the candidate without ranking or scoring.
type CandidateProjection struct {
	Username      string                 `json:"username"`
	DisplayName   string                 `json:"display_name"`
	Company       string                 `json:"company,omitempty"`
	Location      string                 `json:"location,omitempty"`
	Facts         *signals.Facts         `json:"facts,omitempty"`
	Signals       map[string]float64     `json:"signals"`
	Contributions *contributions.Summary `json:"contributions,omitempty"`
}

// BuildCandidateProjection creates a candidate view from a Profile.
// DisplayName falls back to Username if Metadata or Metadata.Name is nil.
func BuildCandidateProjection(p index.Profile) CandidateProjection {
	displayName := p.Username
	company := ""
	location := ""

	if p.Metadata != nil {
		if p.Metadata.Name != "" {
			displayName = p.Metadata.Name
		}
		company = p.Metadata.Company
		location = p.Metadata.Location
	}

	return CandidateProjection{
		Username:      p.Username,
		DisplayName:   displayName,
		Company:       company,
		Location:      location,
		Facts:         p.Facts,
		Signals:       p.Signals,
		Contributions: p.Contributions,
	}
}
