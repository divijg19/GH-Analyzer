package repositoryintelligence

import (
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
)

// Level is a deterministic qualitative classification for a repository dimension.
type Level string

const (
	LevelUnknown   Level = "unknown"
	LevelNone      Level = "none"
	LevelLow       Level = "low"
	LevelModerate  Level = "moderate"
	LevelHigh      Level = "high"
	LevelNarrow    Level = "narrow"
	LevelBroad     Level = "broad"
	LevelEmpty     Level = "empty"
	LevelActive    Level = "active"
	LevelDormant   Level = "dormant"
	LevelInactive  Level = "inactive"
	LevelMaintained Level = "maintained"
	LevelIndividual Level = "individual"
	LevelGoverned  Level = "governed"
	LevelEarly     Level = "early"
	LevelMature    Level = "mature"
	LevelDisciplined Level = "disciplined"
	LevelUndisciplined Level = "undisciplined"
	LevelAtRisk    Level = "at-risk"
	LevelSound     Level = "sound"
)

// Confidence classifies how much evidence supported a dimension's conclusion.
type Confidence string

const (
	ConfidenceLow      Confidence = "low"
	ConfidenceModerate Confidence = "moderate"
	ConfidenceHigh     Confidence = "high"
)

// RepositoryDimension is the normative contract every repository dimension
// satisfies. See docs/REPOSITORY_INTELLIGENCE.md §4.
type RepositoryDimension interface {
	Name() string
	Evidence() []evidence.EvidenceGroup
	Summary() string
}

// dimensionBase provides the normative RepositoryDimension contract. Each
// dimension embeds it and adds its own interpretable fields.
type dimensionBase struct {
	name     string
	evidence []evidence.EvidenceGroup
	summary  string
}

func (d dimensionBase) Name() string                       { return d.name }
func (d dimensionBase) Evidence() []evidence.EvidenceGroup { return d.evidence }
func (d dimensionBase) Summary() string                    { return d.summary }

// RepositoryIntelligence is the deterministic semantic interpretation of a single
// repository. It never introduces information; it reorganizes observable
// repository facts into thirteen interpretable dimensions.
type RepositoryIntelligence struct {
	Repository    string
	ReferenceTime time.Time

	Identity      IdentityDimension
	Health        HealthDimension
	Maintenance   MaintenanceDimension
	Delivery      DeliveryDimension
	Architecture  ArchitectureDimension
	Technology    TechnologyDimension
	Community     CommunityDimension
	Documentation DocumentationDimension
	Quality       QualityDimension
	Complexity    ComplexityDimension
	Lifecycle     LifecycleDimension
	Governance    GovernanceDimension
	Risk          RiskDimension
}

// Dimensions returns every dimension in canonical order.
func (r *RepositoryIntelligence) Dimensions() []RepositoryDimension {
	return []RepositoryDimension{
		r.Identity,
		r.Health,
		r.Maintenance,
		r.Delivery,
		r.Architecture,
		r.Technology,
		r.Community,
		r.Documentation,
		r.Quality,
		r.Complexity,
		r.Lifecycle,
		r.Governance,
		r.Risk,
	}
}
