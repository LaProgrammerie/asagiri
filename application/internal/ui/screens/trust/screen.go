package trust

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains trust explorer data.
type ViewModel struct {
	Trust   bus.TrustExplorerResult
	ShowCLI bool
}

// Render returns trust explorer textual content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Trust Summary\n")
	b.WriteString(fmt.Sprintf("Overall: %.0f%%  Residual risk: %s\n", vm.Trust.Overall*100, value(vm.Trust.ResidualRisk, "unknown")))
	b.WriteString(fmt.Sprintf("Gate: %s\n", value(vm.Trust.GateStatus, "unknown")))
	if vm.Trust.GateReason != "" {
		b.WriteString("Gate reason: " + vm.Trust.GateReason + "\n")
	}
	if vm.Trust.Warning != "" {
		b.WriteString("Warning: " + vm.Trust.Warning + "\n")
	}
	if len(vm.Trust.Dimensions) == 0 {
		b.WriteString("- unavailable")
		return b.String()
	}
	for _, dim := range vm.Trust.Dimensions {
		b.WriteString(fmt.Sprintf("- %s %.0f%%\n", dim.Label, dim.Score*100))
		if len(dim.Findings) > 0 {
			b.WriteString("  finding: " + dim.Findings[0] + "\n")
		}
		if len(dim.Evidence) > 0 {
			b.WriteString("  evidence: " + dim.Evidence[0] + "\n")
		}
		if vm.ShowCLI && dim.CLIEquivalent != "" {
			b.WriteString("  CLI: " + dim.CLIEquivalent + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
