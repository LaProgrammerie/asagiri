package app

import "strings"

// useRichLayout is true when the Experience Platform should render panelised
// layouts (cockpit, dashboard grid). Plain/json modes keep flat text parity
// (FR-2.4 / CK-1.3).
func useRichLayout(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "plain", "json":
		return false
	default:
		return true
	}
}

func (m model) useRichLayout() bool {
	return useRichLayout(m.cfg.Mode)
}
