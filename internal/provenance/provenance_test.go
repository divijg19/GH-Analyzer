package provenance

import (
	"reflect"
	"testing"
)

func TestChainIsEmpty(t *testing.T) {
	if !(Chain{}).IsEmpty() {
		t.Fatal("zero chain should be empty")
	}
	c := Chain{Facts: Facts("total_repos")}
	if c.IsEmpty() {
		t.Fatal("chain with facts should not be empty")
	}
}

func TestRepositoryFieldIdentityAndSource(t *testing.T) {
	ref := RepositoryField("atlas", "Stars")
	if ref.Kind != ObservationRepository || ref.ID != "atlas" || ref.Field != "Stars" || ref.Source != SourceGitHub {
		t.Fatalf("unexpected observation ref: %+v", ref)
	}
}

func TestRepositoryFamilyFieldOmitsID(t *testing.T) {
	ref := RepositoryFamilyField("Size")
	if ref.ID != "" {
		t.Fatalf("family field must omit ID, got %q", ref.ID)
	}
}

func TestMergePreservesOrderDeterministically(t *testing.T) {
	a := Chain{Indicators: Indicators("ownership"), Facts: Facts("valid_repos")}
	b := Chain{Observations: RepositoryObservations("atlas", "Size", "Fork")}

	first := Merge(a, b)
	for i := 0; i < 1000; i++ {
		if got := Merge(a, b); !reflect.DeepEqual(first, got) {
			t.Fatalf("Merge not deterministic on iteration %d", i)
		}
	}

	want := Chain{
		Indicators:   Indicators("ownership"),
		Facts:        Facts("valid_repos"),
		Observations: RepositoryObservations("atlas", "Size", "Fork"),
	}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("Merge = %+v, want %+v", first, want)
	}
}
