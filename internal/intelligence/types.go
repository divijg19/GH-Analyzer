package intelligence

import (
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
)

// Level is a deterministic qualitative classification for an intelligence dimension.
type Level string

const (
	LevelUnknown    Level = "unknown"
	LevelLow        Level = "low"
	LevelModerate   Level = "moderate"
	LevelHigh       Level = "high"
	LevelNarrow     Level = "narrow"
	LevelBroad      Level = "broad"
	LevelInactive   Level = "inactive"
	LevelActive     Level = "active"
	LevelAbandoned  Level = "abandoned"
	LevelMaintained Level = "maintained"
	LevelIndividual Level = "individual"
	LevelEmpty      Level = "empty"
)

// Confidence classifies how much evidence supported a dimension's conclusion.
type Confidence string

const (
	ConfidenceLow      Confidence = "low"
	ConfidenceModerate Confidence = "moderate"
	ConfidenceHigh     Confidence = "high"
)

// Dimension is the normative contract every intelligence dimension satisfies.
// See docs/CANDIDATE_INTELLIGENCE.md §8.
type Dimension interface {
	Name() string
	Evidence() []evidence.EvidenceGroup
	Summary() string
}

// dimensionBase provides the normative Dimension contract for concrete
// dimensions. Each dimension embeds it and adds its own interpretable fields.
type dimensionBase struct {
	name     string
	evidence []evidence.EvidenceGroup
	summary  string
}

func (d dimensionBase) Name() string                       { return d.name }
func (d dimensionBase) Evidence() []evidence.EvidenceGroup { return d.evidence }
func (d dimensionBase) Summary() string                    { return d.summary }

// CandidateIntelligence is the deterministic semantic interpretation of a
// Profile. It never introduces information; it aggregates the per-repository
// RepositoryIntelligence of the Profile into nine interpretable portfolio
// dimensions. Growth is a reserved dimension and is intentionally absent from
// v0.9.0.
type CandidateIntelligence struct {
	Username      string
	ReferenceTime time.Time

	Ownership     OwnershipIntelligence
	Delivery      DeliveryIntelligence
	Breadth       BreadthIntelligence
	Focus         FocusIntelligence
	Maintenance   MaintenanceIntelligence
	Project       ProjectIntelligence
	Collaboration CollaborationIntelligence
	Documentation DocumentationIntelligence
	Portfolio     PortfolioIntelligence
	Topology      TopologyIntelligence
	Technology    TechnologyIntelligence

	// Growth is a reserved dimension. It is intentionally absent from v0.9.0:
	// Atlas lacks historical candidate snapshots required to implement it
	// without a weak proxy. Implemented in v0.10.x+ once snapshot storage
	// exists. See docs/CANDIDATE_INTELLIGENCE.md.
	// Growth *GrowthIntelligence
}

// Dimensions returns every implemented dimension in canonical order.
func (c *CandidateIntelligence) Dimensions() []Dimension {
	return []Dimension{
		c.Ownership,
		c.Delivery,
		c.Breadth,
		c.Focus,
		c.Maintenance,
		c.Project,
		c.Collaboration,
		c.Documentation,
		c.Portfolio,
		c.Topology,
		c.Technology,
	}
}
