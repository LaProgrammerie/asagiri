package onboarding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// AgentKeys returns sorted keys from config.agents (empty when unknown).
func AgentKeys(cfg *config.Config) []string {
	if cfg == nil || len(cfg.Agents) == 0 {
		return nil
	}
	keys := make([]string, 0, len(cfg.Agents))
	for k := range cfg.Agents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func validateAgentRef(value string, known []string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "clé agents.* requise"
	}
	if len(known) == 0 {
		return ""
	}
	for _, k := range known {
		if k == value {
			return ""
		}
	}
	return fmt.Sprintf("%q absent de config.agents", value)
}
