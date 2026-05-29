package confidence

// Normalizer clamps and scales raw scores into [0, 1].
type Normalizer interface {
	Normalize(score float64) float64
}

// ClampNormalizer bounds scores to [0, 1].
type ClampNormalizer struct{}

// Normalize implements Normalizer.
func (ClampNormalizer) Normalize(score float64) float64 {
	return Clamp01(score)
}

// Clamp01 returns score bounded to [0, 1].
func Clamp01(score float64) float64 {
	switch {
	case score < 0:
		return 0
	case score > 1:
		return 1
	default:
		return score
	}
}
