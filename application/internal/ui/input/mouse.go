package input

// MouseConfig controls mouse support.
type MouseConfig struct {
	Enabled bool
}

// DefaultMouseConfig returns default mouse behavior.
func DefaultMouseConfig() MouseConfig {
	return MouseConfig{Enabled: true}
}
