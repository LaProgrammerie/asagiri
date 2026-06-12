package components

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// ReadinessBannerViewModel drives the Mission Control readiness strip.
type ReadinessBannerViewModel struct {
	Ready bool
	Score int
	Theme theme.Theme
}

// ReadinessBanner renders a compact readiness warning when the project is not ready.
func ReadinessBanner(vm ReadinessBannerViewModel) string {
	if vm.Ready {
		return ""
	}
	th := vm.Theme
	if th.Name == "" {
		th = theme.Default()
	}
	st := th.Styles()
	score := vm.Score
	if score < 0 {
		score = 0
	}
	return st.Warning.Render(fmt.Sprintf("⚠ Projet non prêt (%d%%)", score)) + " " +
		st.Muted.Render("asa onboard --yes  |  asa ready --plain  |  asa doctor --full")
}

// ReadinessBannerFromResult builds a banner view model from a bus result.
func ReadinessBannerFromResult(r bus.ReadinessResult) ReadinessBannerViewModel {
	return ReadinessBannerViewModel{Ready: r.Ready, Score: r.Score}
}
