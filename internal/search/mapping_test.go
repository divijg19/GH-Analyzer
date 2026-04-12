package search

import "testing"

func TestMapIntentBackend(t *testing.T) {
	query, err := MapIntent("backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(query.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(query.Conditions))
	}

	condition := query.Conditions[0]
	if condition.Signal != "depth" || condition.Operator != ">=" || condition.Value != 0.6 {
		t.Fatalf("unexpected condition: %+v", condition)
	}
}

func TestMapIntentConsistent(t *testing.T) {
	query, err := MapIntent("consistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(query.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(query.Conditions))
	}

	condition := query.Conditions[0]
	if condition.Signal != "consistency" || condition.Operator != ">=" || condition.Value != 0.7 {
		t.Fatalf("unexpected condition: %+v", condition)
	}
}

func TestMapIntentExpressionFallback(t *testing.T) {
	query, err := MapIntent("consistency > 0.7 AND depth > 0.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(query.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(query.Conditions))
	}
}
