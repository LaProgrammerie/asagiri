package runtime

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Mode describes runtime behaviour (spec-my-A §24.17).
type Mode string

const (
	ModeGuided      Mode = config.RuntimeModeGuided
	ModeInteractive Mode = config.RuntimeModeInteractive
	ModeHeadless    Mode = config.RuntimeModeHeadless
	ModeCI          Mode = config.RuntimeModeCI
	ModeReview      Mode = config.RuntimeModeReview
	ModeExploration Mode = config.RuntimeModeExploration
)

// ParseMode validates a runtime mode string.
func ParseMode(s string) (Mode, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	for _, m := range config.ValidRuntimeModes {
		if s == m {
			return Mode(s), nil
		}
	}
	return "", fmt.Errorf("runtime: invalid mode %q (want one of %v)", s, config.ValidRuntimeModes)
}

// Behaviour hints for orchestration layers.
type Behaviour struct {
	RequireConfirmation bool
	AutoInvestigate     bool
	VerboseEvents       bool
	AllowCloud          bool
}

// BehaviourFor returns orchestration hints per mode.
func BehaviourFor(m Mode) Behaviour {
	switch m {
	case ModeHeadless, ModeCI:
		return Behaviour{RequireConfirmation: false, AutoInvestigate: true, VerboseEvents: false, AllowCloud: false}
	case ModeReview:
		return Behaviour{RequireConfirmation: true, AutoInvestigate: false, VerboseEvents: true, AllowCloud: false}
	case ModeExploration:
		return Behaviour{RequireConfirmation: false, AutoInvestigate: true, VerboseEvents: true, AllowCloud: true}
	case ModeInteractive:
		return Behaviour{RequireConfirmation: true, AutoInvestigate: true, VerboseEvents: true, AllowCloud: true}
	default: // guided
		return Behaviour{RequireConfirmation: true, AutoInvestigate: false, VerboseEvents: true, AllowCloud: false}
	}
}
