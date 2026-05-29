package knowledge

import "fmt"

// GraphSource records where a node or edge was derived from (spec-my-E §20).
type GraphSource struct {
	Kind      string `json:"kind,omitempty"`
	Path      string `json:"path,omitempty"`
	Extractor string `json:"extractor,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
}

// Validate checks that provenance is sufficient for persistence.
func (s GraphSource) Validate() error {
	if s.Kind == "" && s.Path == "" && s.Extractor == "" {
		return fmt.Errorf("%w: source kind, path, or extractor required", ErrInvalidProvenance)
	}
	return nil
}

// ValidateConfidence ensures confidence is in [0, 1].
func ValidateConfidence(confidence float64) error {
	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("%w: confidence must be between 0 and 1", ErrInvalidProvenance)
	}
	return nil
}

// ValidateUpsertProvenance requires source and confidence on writes.
func ValidateUpsertProvenance(source GraphSource, confidence float64) error {
	if err := source.Validate(); err != nil {
		return err
	}
	return ValidateConfidence(confidence)
}
