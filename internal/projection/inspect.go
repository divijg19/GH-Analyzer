package projection

import (
	"github.com/divijg19/Atlas/internal/contributions"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/profile"
	"github.com/divijg19/Atlas/internal/signals"
)

// InspectProjection provides a raw data view for inspection and debugging.
// Contains everything from Profile plus generated evidence.
type InspectProjection struct {
	Username      string                   `json:"username"`
	Metadata      *profile.UserMetadata    `json:"metadata,omitempty"`
	Facts         *signals.RepositoryFacts `json:"facts,omitempty"`
	Signals       map[string]float64       `json:"signals"`
	Contributions *contributions.Summary   `json:"contributions,omitempty"`
	Evidence      []signals.EvidenceGroup  `json:"evidence"`
}

// BuildInspectProjection creates a raw data view from a Profile.
// Generates evidence from facts and signals if available.
func BuildInspectProjection(p index.Profile) InspectProjection {
	var evidence []signals.EvidenceGroup

	if p.Facts != nil {
		sig := signals.FromMap(p.Signals)
		evidence = signals.GenerateEvidence(*p.Facts, sig)
	}

	return InspectProjection{
		Username:      p.Username,
		Metadata:      p.Metadata,
		Facts:         p.Facts,
		Signals:       p.Signals,
		Contributions: p.Contributions,
		Evidence:      evidence,
	}
}
