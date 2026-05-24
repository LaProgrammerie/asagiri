package intent

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrIntentDisabled = errors.New("intent layer désactivée (intent.enabled=false)")
	ErrLowConfidence  = errors.New("intention ambiguë")
)

// AmbiguityError is returned in non-interactive mode when confidence is low.
type AmbiguityError struct {
	Instruction string
	Candidates  []string
	Confidence  float64
}

func (e *AmbiguityError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "instruction ambiguë (confiance %.2f)", e.Confidence)
	if len(e.Candidates) > 0 {
		b.WriteString("\nDid you mean:\n")
		for i, c := range e.Candidates {
			fmt.Fprintf(&b, "%d. %s\n", i+1, c)
		}
	}
	return strings.TrimSpace(b.String())
}

func (e *AmbiguityError) Is(target error) bool {
	return target == ErrLowConfidence
}

// ConfirmationRequiredError signals guided mode needs user approval.
type ConfirmationRequiredError struct {
	Message string
}

func (e *ConfirmationRequiredError) Error() string {
	return e.Message
}
