package projection

import (
	"github.com/divijg19/Atlas/internal/contributions"
	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/indicators"
	"github.com/divijg19/Atlas/internal/profile"
)

// InspectProjection provides a raw data view for inspection and debugging.
// Contains everything from Profile plus generated evidence.
type InspectProjection struct {
	Username      string                   `json:"username"`
	Metadata      *profile.UserMetadata    `json:"metadata,omitempty"`
	Facts         *facts.RepositoryFacts   `json:"facts,omitempty"`
	Signals       map[string]float64       `json:"signals"`
	Contributions *contributions.Summary   `json:"contributions,omitempty"`
	ActivityFacts *facts.ActivityFacts     `json:"activity_facts,omitempty"`
	Evidence      []evidence.EvidenceGroup `json:"evidence"`
}

// BuildInspectProjection creates a raw data view from a Profile.
// Generates evidence from facts and indicators if available.
func BuildInspectProjection(p index.Profile) InspectProjection {
	var groupedEvidence []evidence.EvidenceGroup

	if p.Facts != nil {
		sig := indicators.FromMap(p.Signals)
		groupedEvidence = evidence.GenerateEvidence(*p.Facts, sig)
	}

	return InspectProjection{
		Username:      p.Username,
		Metadata:      p.Metadata,
		Facts:         p.Facts,
		Signals:       p.Signals,
		Contributions: p.Contributions,
		ActivityFacts: p.ActivityFacts,
		Evidence:      groupedEvidence,
	}
}
