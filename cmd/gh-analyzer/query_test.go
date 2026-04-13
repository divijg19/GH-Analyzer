package main

import (
	"strings"
	"testing"
)

func TestParseConditionInvalidSignal(t *testing.T) {
	_, err := parseCondition("unknown>=0.5")
	if err == nil {
		t.Fatal("expected error for invalid signal")
	}
}

func TestParseConditionInvalidOperator(t *testing.T) {
	_, err := parseCondition("consistency==0.5")
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
}

func TestParseExpressionMalformed(t *testing.T) {
	_, err := parseExpression("consistency>=0.7 AND ")
	if err == nil {
		t.Fatal("expected error for malformed expression")
	}
}

func TestConditionsFromPreset(t *testing.T) {
	conditions, err := conditionsFromPreset("strong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(conditions))
	}
}

func TestMissingDatasetError(t *testing.T) {
	err := missingDatasetError(defaultDatasetPath)
	if err == nil {
		t.Fatal("expected error")
	}

	message := err.Error()
	if !strings.Contains(message, "dataset not found") {
		t.Fatalf("expected dataset-not-found message, got %q", message)
	}
	if !strings.Contains(message, "Run: gh-analyzer build <usernames>") {
		t.Fatalf("expected actionable guidance, got %q", message)
	}
}
