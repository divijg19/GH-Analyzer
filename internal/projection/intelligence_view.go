package projection

import (
	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/intelligence"
)

// DimensionView is a presentation-oriented view of a single intelligence
// dimension. Projection selects which fields to emphasize (Analyze: summary;
// Inspect: evidence); it never recomputes meaning. The canonical data lives in
// intelligence.CandidateIntelligence.
type DimensionView struct {
	Name     string                   `json:"name"`
	Level    string                   `json:"level"`
	Summary  string                   `json:"summary"`
	Evidence []evidence.EvidenceGroup `json:"evidence,omitempty"`
}

// IntelligenceView builds presentation views from a CandidateIntelligence. It
// reads the concrete dimension types so the Level is preserved; Summary and
// Evidence come from the normative Dimension contract. Projection never
// derives or reinterprets these values.
func IntelligenceView(ci *intelligence.CandidateIntelligence) []DimensionView {
	if ci == nil {
		return nil
	}
	return []DimensionView{
		dimView("ownership", ci.Ownership.Level, ci.Ownership),
		dimView("delivery", ci.Delivery.Level, ci.Delivery),
		dimView("breadth", ci.Breadth.Level, ci.Breadth),
		dimView("focus", ci.Focus.Level, ci.Focus),
		dimView("maintenance", ci.Maintenance.Level, ci.Maintenance),
		dimView("project", ci.Project.Level, ci.Project),
		dimView("collaboration", ci.Collaboration.Level, ci.Collaboration),
		dimView("documentation", ci.Documentation.Level, ci.Documentation),
		dimView("portfolio", ci.Portfolio.Level, ci.Portfolio),
	}
}

func dimView(name string, level intelligence.Level, d intelligence.Dimension) DimensionView {
	return DimensionView{
		Name:     name,
		Level:    string(level),
		Summary:  d.Summary(),
		Evidence: d.Evidence(),
	}
}
