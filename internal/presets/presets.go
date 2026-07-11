package presets

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/signals"
)

func Preset(name string) (engine.Query, error) {
	switch name {
	case "strong":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: signals.SignalConsistency, Operator: ">=", Value: 0.7},
				{Signal: signals.SignalOwnership, Operator: ">=", Value: 0.6},
			},
		}, nil
	case "consistent":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: signals.SignalConsistency, Operator: ">=", Value: 0.8},
			},
		}, nil
	case "deep":
		return engine.Query{
			Conditions: []engine.Condition{
				{Signal: signals.SignalDepth, Operator: ">=", Value: 0.7},
			},
		}, nil
	default:
		return engine.Query{}, fmt.Errorf("unknown preset %q", name)
	}
}
