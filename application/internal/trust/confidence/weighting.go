package confidence

// Weighter returns the relative weight of a trust dimension in the overall score.
type Weighter interface {
	Weight(d Dimension) float64
}

// DefaultWeighter uses equal weights across six dimensions (spec §7).
type DefaultWeighter struct {
	weights map[Dimension]float64
}

// NewDefaultWeighter returns a weighter with DefaultWeights.
func NewDefaultWeighter() DefaultWeighter {
	return DefaultWeighter{weights: DefaultWeights()}
}

// Weight implements Weighter.
func (w DefaultWeighter) Weight(d Dimension) float64 {
	if w.weights == nil {
		w.weights = DefaultWeights()
	}
	return w.weights[d]
}

// DefaultWeights assigns equal weight across the six dimensions (spec §7).
func DefaultWeights() map[Dimension]float64 {
	return map[Dimension]float64{
		DimensionArchitecture:   1.0 / 6.0,
		DimensionImplementation: 1.0 / 6.0,
		DimensionFlowIntegrity:  1.0 / 6.0,
		DimensionObservability:  1.0 / 6.0,
		DimensionSecurity:       1.0 / 6.0,
		DimensionRegression:     1.0 / 6.0,
	}
}
