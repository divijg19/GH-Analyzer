// Package projection provides deterministic, read-only views of GH-Analyzer domain objects.
//
// This package owns NO business logic, NO persistence, and NO side effects. It exists
// solely to project internal domain objects (Profile) into stable, consumable representations
// (CandidateProjection) for presentation layers (CLI, API, exports).
//
// Architectural Rules:
//   - CandidateProjection is derived from Profile, never the reverse.
//   - No domain package imports this package.
//   - Projections are views, not sources of truth.
//   - All fields are pure data with no formatting or presentation logic.
package projection

import (
	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

// CandidateProjection is a canonical, immutable view of a GitHub candidate.
// It derives exclusively from Profile and presents data for uniform consumption
// by presentation layers (CLI, API, exports, future Lattice integration).
type CandidateProjection struct {
	Username      string
	DisplayName   string
	Company       string
	Location      string
	Facts         *signals.Facts
	Signals       map[string]float64
	Contributions *contributions.Summary
}

// BuildCandidateProjection projects a Profile into a deterministic CandidateProjection.
// Pure function: no mutation, no side effects, no business logic.
// DisplayName falls back to Username if Metadata.Name is empty.
// Facts, Signals, and Contributions are shallow references to Profile data.
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
