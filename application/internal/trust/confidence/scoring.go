package confidence

import (
	"context"
	"strings"
)

// DetailedCheck carries per-check data for real aggregation (lot 2).
type DetailedCheck struct {
	Type       string
	Status     string
	Confidence float64
	Findings   []FindingInput
}

const checkStatusSkipped = "skipped"

// InferredDimensionCap bounds observability/security without a dedicated check (lot 2).
const InferredDimensionCap = 0.5

// InferredDimensions have no dedicated lot-2 check; scores are capped for reporting.
var InferredDimensions = []Dimension{DimensionObservability, DimensionSecurity}

// FindingInput is a severity/category pair for scoring penalties.
type FindingInput struct {
	Severity string
	Category string
}

// Scorer derives per-dimension raw scores from check contributions.
type Scorer interface {
	Score(ctx context.Context, checks []DetailedCheck) (map[Dimension]float64, error)
}

// DefaultScorer maps lot-2 checks and findings to dimension raw scores.
type DefaultScorer struct{}

var checkDimensionMatrix = map[string]map[Dimension]float64{
	"static-analysis": {
		DimensionArchitecture:   0.2,
		DimensionImplementation: 0.5,
		DimensionRegression:     0.3,
	},
	"contracts": {
		DimensionArchitecture:   0.6,
		DimensionImplementation: 0.1,
		DimensionFlowIntegrity:  0.1,
		DimensionObservability:  0.1,
		DimensionSecurity:       0.1,
	},
	"flows": {
		DimensionArchitecture:  0.1,
		DimensionFlowIntegrity: 0.5,
		DimensionObservability: 0.15,
		DimensionSecurity:      0.15,
		DimensionRegression:    0.1,
	},
	"tests": {
		DimensionImplementation: 0.4,
		DimensionRegression:     0.6,
	},
	"blast-radius": {
		DimensionRegression:   0.7,
		DimensionArchitecture: 0.3,
	},
	"observability": {
		DimensionObservability: 0.8,
		DimensionFlowIntegrity: 0.2,
	},
	"security": {
		DimensionSecurity:      0.85,
		DimensionFlowIntegrity: 0.15,
	},
	"permissions": {
		DimensionSecurity:      0.5,
		DimensionFlowIntegrity: 0.5,
	},
	"cost": {
		DimensionRegression:   0.3,
		DimensionArchitecture: 0.7,
	},
	"analytics": {
		DimensionObservability: 0.4,
		DimensionArchitecture:  0.6,
	},
	"architecture": {
		DimensionArchitecture:   0.9,
		DimensionImplementation: 0.1,
	},
	"performance": {
		DimensionImplementation: 0.7,
		DimensionRegression:     0.3,
	},
	"backward-compatibility": {
		DimensionRegression:   0.8,
		DimensionArchitecture: 0.2,
	},
	"migration-safety": {
		DimensionRegression:   0.6,
		DimensionArchitecture: 0.4,
	},
}

var severityWeight = map[string]float64{
	"critical": 0.35,
	"error":    0.20,
	"warning":  0.08,
	"info":     0.02,
}

var categoryDimensions = map[string][]Dimension{
	"architecture.contract":   {DimensionArchitecture},
	"architecture.dependency": {DimensionArchitecture, DimensionRegression},
	"architecture.flow":       {DimensionArchitecture},
	"implementation.static":   {DimensionImplementation},
	"implementation.test":     {DimensionImplementation, DimensionRegression},
	"flow.integrity":          {DimensionFlowIntegrity},
	"flow.observability":      {DimensionFlowIntegrity, DimensionObservability},
	"flow.security":           {DimensionFlowIntegrity, DimensionSecurity},
	"contract.openapi":        {DimensionArchitecture},
	"regression.test":         {DimensionRegression},
	"observability.contract":  {DimensionObservability},
	"observability.flow":      {DimensionObservability},
	"security.flow":           {DimensionSecurity},
	"security.contract":       {DimensionSecurity},
	"permissions.flow":        {DimensionSecurity, DimensionFlowIntegrity},
	"cost.flow":               {DimensionRegression},
	"analytics.contract":      {DimensionObservability, DimensionArchitecture},
	"performance.static":      {DimensionImplementation},
	"compatibility.contract":  {DimensionRegression},
	"migration.dependency":    {DimensionRegression},
	"blast.radius":            {DimensionRegression},
}

// Score implements Scorer.
func (DefaultScorer) Score(_ context.Context, checks []DetailedCheck) (map[Dimension]float64, error) {
	sum := make(map[Dimension]float64)
	weightSum := make(map[Dimension]float64)
	penalties := make(map[Dimension]float64)

	for _, c := range checks {
		if c.Status == checkStatusSkipped {
			continue
		}
		matrix, ok := checkDimensionMatrix[c.Type]
		if !ok {
			continue
		}
		for d, w := range matrix {
			sum[d] += c.Confidence * w
			weightSum[d] += w
		}
		for _, f := range c.Findings {
			pen := severityWeight[strings.ToLower(f.Severity)]
			for _, d := range categoryDimensions[f.Category] {
				penalties[d] += pen
			}
		}
	}

	raw := make(map[Dimension]float64)
	for _, d := range AllDimensions {
		if weightSum[d] == 0 {
			continue
		}
		raw[d] = Clamp01(sum[d]/weightSum[d] - penalties[d])
	}
	return raw, nil
}
