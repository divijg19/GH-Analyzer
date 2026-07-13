package engine

import (
	"strings"

	"github.com/divijg19/Atlas/internal/index"
)

type Query struct {
	Conditions []Condition
	Limit      int
}

type Condition struct {
	Signal   string
	Operator string
	Value    float64
}

func match(p index.Profile, c Condition) bool {
	signal := strings.ToLower(strings.TrimSpace(c.Signal))
	signalValue, ok := p.Signals[signal]
	if !ok {
		return false
	}

	switch strings.TrimSpace(c.Operator) {
	case ">":
		return signalValue > c.Value
	case ">=":
		return signalValue >= c.Value
	case "<":
		return signalValue < c.Value
	case "<=":
		return signalValue <= c.Value
	default:
		return false
	}
}
