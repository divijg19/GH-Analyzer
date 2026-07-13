// Package presets owns named search presets consumed by the engine layer.
//
// It maps human-readable preset names ("strong", "consistent", "deep") to
// predefined engine.Query values.
//
// Presets never acquire observations, derive facts, compute indicators,
// evaluate candidates, or perform presentation.
//
// Consumed by: search, cmd/atlas.
package presets

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/indicators"
)

func Preset(name string) (engine.Query, error) {
	switch name {
	case "strong":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: indicators.SignalConsistency, Operator: ">=", Value: 0.7},
				{Signal: indicators.SignalOwnership, Operator: ">=", Value: 0.6},
			},
		}, nil
	case "consistent":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: indicators.SignalConsistency, Operator: ">=", Value: 0.8},
			},
		}, nil
	case "deep":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: indicators.SignalDepth, Operator: ">=", Value: 0.7},
			},
		}, nil
	default:
		return engine.Query{}, fmt.Errorf("unknown preset %q", name)
	}
}
