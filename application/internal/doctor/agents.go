package doctor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func collectAgents(cfg *config.Config) ([]AgentInfo, []MissingToolInfo) {
	if cfg == nil {
		return nil, nil
	}
	roles := []struct {
		role string
		id   string
	}{
		{"spec", cfg.Work.DefaultSpecAgent},
		{"enricher", cfg.Work.DefaultEnricher},
		{"dev", cfg.Work.DefaultAgent},
		{"reviewer", cfg.Work.DefaultReviewer},
	}
	var agents []AgentInfo
	var missing []MissingToolInfo
	for _, r := range roles {
		info, miss := agentInfo(cfg, r.role, strings.TrimSpace(r.id))
		agents = append(agents, info)
		if miss != nil {
			missing = append(missing, *miss)
		}
	}
	return agents, missing
}

func agentInfo(cfg *config.Config, role, logicalID string) (AgentInfo, *MissingToolInfo) {
	info := AgentInfo{Role: role, LogicalID: logicalID}
	if logicalID == "" {
		info.Status = "warn"
		info.Detail = "non défini dans config.work"
		return info, &MissingToolInfo{
			Name:   "work." + role,
			Reason: info.Detail,
			FixCLI: "asa onboard --step agents",
		}
	}
	agent, ok := cfg.Agents[logicalID]
	if !ok {
		info.Status = "missing"
		info.Detail = fmt.Sprintf("agents.%s absent", logicalID)
		return info, &MissingToolInfo{Name: logicalID, Reason: info.Detail, FixCLI: "asa onboard --step agents"}
	}
	providerType, merged, err := cfg.MergedAgentRuntime(logicalID)
	if err != nil {
		info.Status = "warn"
		info.Detail = err.Error()
		return info, nil
	}
	info.Provider = strings.TrimSpace(agent.Provider)
	if info.Provider == "" {
		info.Provider = strings.TrimSpace(providerType)
	}
	info.Command = strings.TrimSpace(merged.Command)
	if info.Command == "" && strings.TrimSpace(merged.Endpoint) != "" {
		info.Status = "ok"
		info.InPath = true
		info.Detail = "API " + merged.Endpoint
		return info, nil
	}
	if info.Command == "" {
		info.Status = "warn"
		info.Detail = "command manquante"
		return info, &MissingToolInfo{
			Name:   logicalID,
			Reason: info.Detail,
			FixCLI: "Éditer .asagiri/config.yaml (agents ou providers)",
		}
	}
	if _, err := exec.LookPath(info.Command); err != nil {
		info.Status = "missing"
		info.Detail = fmt.Sprintf("%q introuvable dans PATH", info.Command)
		return info, &MissingToolInfo{Name: info.Command, Reason: info.Detail, FixCLI: "Installer l'outil ou corriger config"}
	}
	info.Status = "ok"
	info.InPath = true
	return info, nil
}
