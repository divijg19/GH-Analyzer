package projection

import (
	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/profile"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

// InspectProjection provides a raw data view for inspection and debugging.
// Contains everything from Profile plus generated evidence.
type InspectProjection struct {
	Username      string                  `json:"username"`
	Metadata      *profile.UserMetadata   `json:"metadata,omitempty"`
	Facts         *signals.Facts          `json:"facts,omitempty"`
	Signals       map[string]float64      `json:"signals"`
	Contributions *contributions.Summary  `json:"contributions,omitempty"`
	Evidence      []signals.EvidenceGroup `json:"evidence"`
}

// BuildInspectProjection creates a raw data view from a Profile.
// Generates evidence from facts and signals if available.
func BuildInspectProjection(p index.Profile) InspectProjection {
	var evidence []signals.EvidenceGroup

	if p.Facts != nil {
		sig := signals.Signals{
			Ownership:   p.Signals["ownership"],
			Consistency: p.Signals["consistency"],
			Depth:       p.Signals["depth"],
			Activity:    p.Signals["activity"],
		}
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
