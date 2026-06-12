package agentexternal

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// DriftInfo describes external provider profile drift for doctor integration.
type DriftInfo struct {
	Kind    string
	Message string
	FixCLI  string
}

// ExternalDrift inspects one agent for external profile drift (read-only).
func ExternalDrift(spec agentspec.Spec, cfg *config.Config, configKey string) (DriftInfo, bool) {
	home, _ := os.UserHomeDir()
	target := buildTarget(spec, configKey, cfg, home)
	if !isAllowlistedPathSource(target.PathSource) {
		return DriftInfo{}, false
	}
	fixCLI := fmt.Sprintf("asa agents external sync --write --agent %s", spec.ID)
	switch target.Status {
	case StatusDrift:
		msg := strings.TrimSpace(target.Detail)
		if msg == "" {
			msg = "profil externe désaligné"
		}
		return DriftInfo{Kind: "external_drift", Message: msg, FixCLI: fixCLI}, true
	case StatusMissingFile:
		return DriftInfo{Kind: "external_missing", Message: "fichier profil provider absent", FixCLI: fixCLI}, true
	default:
		return DriftInfo{}, false
	}
}
