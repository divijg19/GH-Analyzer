package indicators

import (
	"reflect"
	"testing"
)

// TestSignalProvenanceCompleteChain verifies every signal resolves to a complete
// deterministic chain: indicator -> facts -> repository observations. This is the
// Observation -> Fact -> Indicator portion of the certification chain.
func TestSignalProvenanceCompleteChain(t *testing.T) {
	for _, signal := range []string{SignalOwnership, SignalConsistency, SignalDepth, SignalActivity} {
		chain := SignalProvenance(signal)

		if len(chain.Indicators) != 1 || chain.Indicators[0].Signal != signal {
			t.Fatalf("%s: indicator ref missing: %+v", signal, chain.Indicators)
		}
		if len(chain.Facts) == 0 {
			t.Fatalf("%s: no supporting facts", signal)
		}
		if len(chain.Observations) == 0 {
			t.Fatalf("%s: chain does not reach observations", signal)
		}
		for _, o := range chain.Observations {
			if o.Field == "" || o.Source == "" {
				t.Fatalf("%s: observation ref incomplete: %+v", signal, o)
			}
		}
	}
}

func TestSignalProvenanceDeterministic(t *testing.T) {
	first := SignalProvenance(SignalDepth)
	for i := 0; i < 1000; i++ {
		if got := SignalProvenance(SignalDepth); !reflect.DeepEqual(first, got) {
			t.Fatalf("SignalProvenance not deterministic on iteration %d", i)
		}
	}
}
