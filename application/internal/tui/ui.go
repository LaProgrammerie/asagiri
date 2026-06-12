package tui

import (
	"io"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/mattn/go-isatty"
)

// UIMode selects rendering backend (specv3 §13).
type UIMode string

const (
	ModeAuto  UIMode = "auto"
	ModeRich  UIMode = "rich"
	ModePlain UIMode = "plain"
	ModeJSON  UIMode = "json"
)

// UI is a minimal terminal facade (engine must not depend on heavy TUI loops).
type UI interface {
	Box(title, body string)
	ProgressLine(label string, pct float64)
	Event(kind string, payload any)
	Printf(format string, args ...any)
}

// NewUI builds a UI using config.ui.mode when present.
func NewUI(cfg *config.Config, stdout io.Writer, isTTY bool) UI {
	mode := ModeAuto
	if cfg != nil && cfg.UI.Mode != "" {
		mode = UIMode(cfg.UI.Mode)
	}
	return newUIWithMode(mode, stdout, isTTY)
}

func newUIWithMode(mode UIMode, stdout io.Writer, isTTY bool) UI {
	switch effectiveMode(mode, isTTY) {
	case ModeJSON:
		return newJSONUI(stdout)
	case ModeRich:
		if isTTY {
			return newRichUI(stdout)
		}
		fallthrough
	default:
		return newPlainUI(stdout)
	}
}

func effectiveMode(mode UIMode, isTTY bool) UIMode {
	if mode == ModeAuto {
		if isTTY {
			return ModeRich
		}
		return ModePlain
	}
	return mode
}

// DetectTTY is true when w is a terminal.
func DetectTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}
