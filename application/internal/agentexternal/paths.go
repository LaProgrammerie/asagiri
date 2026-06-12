package agentexternal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// allowedPathSources are the only origins permitted for external writes (explicit allowlist).
var allowedPathSources = map[string]struct{}{
	"spec.external.path":          {},
	"config.agents.external_path": {},
}

func isAllowlistedPathSource(source string) bool {
	_, ok := allowedPathSources[strings.TrimSpace(source)]
	return ok
}

func resolveWritableTarget(spec agentspec.Spec, cfg *config.Config, configKey, home string) (ExternalTarget, error) {
	target := buildTarget(spec, configKey, cfg, home)
	if !isAllowlistedPathSource(target.PathSource) {
		target.Status = StatusRejected
		target.Detail = "chemin hors allowlist explicite (spec.external.path ou config.agents.external_path requis)"
		return target, nil
	}
	if target.Status == StatusMissingPath || target.Status == StatusPathUnexpanded {
		target.Status = StatusRejected
		if target.Detail == "" {
			target.Detail = target.Status
		}
		return target, nil
	}
	if target.SupportLevel == string(agentadapter.SupportUnsupported) {
		target.Status = StatusRejected
		if target.Detail == "" {
			target.Detail = "provider non supporté pour cet AgentSpec"
		}
		return target, nil
	}
	if strings.TrimSpace(target.DetectedPath) == "" {
		target.Status = StatusRejected
		target.Detail = "chemin externe non résolu"
		return target, nil
	}
	if !target.Writable {
		target.Status = StatusRejected
		target.Detail = "destination non inscriptible"
		return target, nil
	}
	return target, nil
}

func ensureAllowlistedAbsPath(configuredPath, absPath, source, home string) error {
	if !isAllowlistedPathSource(source) {
		return fmt.Errorf("chemin hors allowlist explicite")
	}
	expected := expandPath(configuredPath, home)
	if expected == "" {
		return fmt.Errorf("chemin externe absent")
	}
	if filepath.Clean(expected) != filepath.Clean(absPath) {
		return fmt.Errorf("chemin résolu %q ≠ allowlist %q", absPath, expected)
	}
	return nil
}
