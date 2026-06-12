package doctor

import (
	"fmt"
	"io"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

// FormatText renders the doctor report for the terminal.
func FormatText(w io.Writer, report Report, strict bool) error {
	writeHeader(w, "Asagiri Doctor")
	writeSection(w, "Repository")
	if report.Repository.GitRoot != "" {
		_, _ = fmt.Fprintf(w, "  Git:    %s\n", report.Repository.GitRoot)
	}
	if report.Repository.ConfigPath != "" {
		status := "absente"
		if report.Repository.ConfigLoaded {
			status = "chargée"
		} else if report.Repository.ConfigError != "" {
			status = "erreur"
		}
		_, _ = fmt.Fprintf(w, "  Config: %s (%s)\n", report.Repository.ConfigPath, status)
		if report.Repository.ConfigError != "" {
			_, _ = fmt.Fprintf(w, "          %s\n", report.Repository.ConfigError)
		}
	}

	writeSection(w, "State")
	if report.State.SQLitePath != "" {
		present := "absent"
		if report.State.SQLitePresent {
			present = "présent"
		}
		_, _ = fmt.Fprintf(w, "  SQLite: %s (%s", report.State.SQLitePath, present)
		if report.State.SchemaVersion > 0 {
			_, _ = fmt.Fprintf(w, ", schema v%d", report.State.SchemaVersion)
		}
		_, _ = fmt.Fprintln(w, ")")
	}
	if report.State.RunCount > 0 || report.State.TaskCount > 0 {
		_, _ = fmt.Fprintf(w, "  Runs: %d  Tasks: %d", report.State.RunCount, report.State.TaskCount)
		if f := strings.TrimSpace(report.State.ActiveFeature); f != "" {
			_, _ = fmt.Fprintf(w, "  Active: %s", f)
		}
		_, _ = fmt.Fprintln(w)
	}

	if len(report.Gates) > 0 {
		writeSection(w, "Gates (work)")
		for _, g := range report.Gates {
			_, _ = fmt.Fprintf(w, "  %-16s %s", g.Name, gateStatusLabel(g))
			if d := strings.TrimSpace(g.Detail); d != "" {
				_, _ = fmt.Fprintf(w, " — %s", d)
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	if len(report.Agents) > 0 {
		writeSection(w, "Agents")
		for _, a := range report.Agents {
			line := fmt.Sprintf("  %-10s %s", a.Role+":", agentLine(a))
			_, _ = fmt.Fprintln(w, line)
		}
	}

	if report.AgentRegistry.Path != "" || len(report.AgentSpecs) > 0 {
		writeSection(w, "Agent specs")
		if report.AgentRegistry.Path != "" {
			reg := report.AgentRegistry
			label := "absent"
			if reg.Present {
				label = fmt.Sprintf("%d fichier(s)", reg.FileCount)
			} else if reg.UsingEmbedded {
				label = "templates embarqués"
			}
			_, _ = fmt.Fprintf(w, "  Registry: %s (%s)\n", agentspecRegistryRel(report.AgentRegistry.Path), label)
			if d := strings.TrimSpace(reg.Detail); d != "" {
				_, _ = fmt.Fprintf(w, "  %s\n", d)
			}
		}
		for _, s := range report.AgentSpecs {
			_, _ = fmt.Fprintf(w, "  %-12s %s\n", s.ConfigKey+":", agentSpecLine(s))
		}
		if report.LastOrchestrated != nil {
			lo := report.LastOrchestrated
			_, _ = fmt.Fprintf(w, "  Dernier contexte: task %s agent %s", lo.TaskID, lo.AgentID)
			if h := truncateHash(lo.AgentHash); h != "" {
				_, _ = fmt.Fprintf(w, " hash %s", h)
			}
			if p := strings.TrimSpace(lo.LogPath); p != "" {
				_, _ = fmt.Fprintf(w, " (%s)", p)
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	if len(report.AgentDrift) > 0 {
		writeSection(w, "Agent drift")
		for _, d := range report.AgentDrift {
			_, _ = fmt.Fprintf(w, "  ⚠ %s — %s\n", d.ConfigKey, d.Message)
			if cli := strings.TrimSpace(d.FixCLI); cli != "" {
				_, _ = fmt.Fprintf(w, "    → %s\n", cli)
			}
		}
	}

	if len(report.MissingTools) > 0 {
		writeSection(w, "Commandes manquantes")
		for _, m := range report.MissingTools {
			_, _ = fmt.Fprintf(w, "  • %s: %s\n", m.Name, m.Reason)
		}
	}

	if report.Trust != nil {
		writeSection(w, "Trust")
		_, _ = fmt.Fprintf(w, "  Feature: %s  Verdict: %s", report.Trust.Feature, report.Trust.Verdict)
		if report.Trust.TasksAtRisk > 0 {
			_, _ = fmt.Fprintf(w, "  (%s)", fmtTasksAtRisk(report.Trust.TasksAtRisk))
		}
		_, _ = fmt.Fprintln(w)
		if s := strings.TrimSpace(report.Trust.Summary); s != "" {
			_, _ = fmt.Fprintf(w, "  %s\n", s)
		}
	}

	if len(report.Warnings) > 0 {
		writeSection(w, "Avertissements")
		for _, c := range report.Warnings {
			_, _ = fmt.Fprintf(w, "  ⚠ %s", c.ID)
			if msg := strings.TrimSpace(c.Message); msg != "" {
				_, _ = fmt.Fprintf(w, ": %s", msg)
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	writeSection(w, "Checks")
	for _, c := range report.Checks {
		_, _ = fmt.Fprintf(w, "  %s %s", checkIcon(c.Status), c.ID)
		if msg := strings.TrimSpace(c.Message); msg != "" {
			_, _ = fmt.Fprintf(w, ": %s", msg)
		}
		_, _ = fmt.Fprintln(w)
	}

	if len(report.NextActions) > 0 {
		writeSection(w, "Prochaines actions")
		for _, a := range report.NextActions {
			title := strings.TrimSpace(a.Title)
			if title == "" {
				title = "Action suggérée"
			}
			_, _ = fmt.Fprintf(w, "  → %s\n     %s\n", a.CLI, title)
		}
	}

	if ShouldFail(report, strict) {
		if len(report.Failures) > 0 {
			return fmt.Errorf("doctor: au moins un contrôle bloquant a échoué")
		}
		return fmt.Errorf("doctor: avertissements non résolus (mode --strict)")
	}
	if len(report.Warnings) > 0 {
		_, err := fmt.Fprintln(w, "Asagiri est prêt (avertissements).")
		return err
	}
	_, err := fmt.Fprintln(w, "Asagiri est prêt.")
	return err
}

func gateStatusLabel(g GateInfo) string {
	switch g.Status {
	case "active":
		return "active"
	case "invalid_mode":
		return fmt.Sprintf("mode invalide (%s)", g.Mode)
	case "inactive":
		return "inactive"
	default:
		return "disabled"
	}
}

func agentspecRegistryRel(absPath string) string {
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return agentspecRegistryDir
	}
	if i := strings.Index(absPath, agentspecRegistryDir); i >= 0 {
		return absPath[i:]
	}
	return absPath
}

const agentspecRegistryDir = ".asagiri/agents"

func agentSpecLine(s AgentSpecEntry) string {
	parts := []string{}
	switch s.PromptSource {
	case "disk":
		parts = append(parts, "spec disque")
		if s.SpecVersion != "" {
			parts = append(parts, "v"+s.SpecVersion)
		}
		if h := truncateHash(s.ContentHash); h != "" {
			parts = append(parts, "hash "+h)
		}
	case "embedded":
		parts = append(parts, "embedded OK")
	case "missing":
		parts = append(parts, "spec manquante")
	case "invalid":
		parts = append(parts, "spec invalide")
	default:
		if s.PromptSource != "" {
			parts = append(parts, s.PromptSource)
		}
	}
	if s.OutputFormat != "" {
		parts = append(parts, "format "+s.OutputFormat)
	}
	if pt := strings.TrimSpace(s.ProviderType); pt != "" {
		support := strings.TrimSpace(s.ProviderSupport)
		if support == "" {
			support = "?"
		}
		parts = append(parts, pt+"→"+support)
	}
	if len(s.Drift) > 0 {
		parts = append(parts, "drift:"+strings.Join(s.Drift, ","))
	}
	switch s.Status {
	case StatusOK:
		if len(parts) == 0 {
			return "ok"
		}
	case StatusWarn:
		parts = append(parts, "⚠")
	case StatusFail:
		parts = append(parts, "✗")
	}
	if d := strings.TrimSpace(s.Detail); d != "" && s.Status != StatusOK {
		parts = append(parts, d)
	}
	return strings.Join(parts, " ")
}

func agentLine(a AgentInfo) string {
	parts := []string{}
	if a.LogicalID != "" {
		parts = append(parts, a.LogicalID)
	}
	switch a.Status {
	case "ok":
		if a.Command != "" {
			parts = append(parts, "✓ "+a.Command)
		} else if a.Detail != "" {
			parts = append(parts, "✓ "+a.Detail)
		}
	case "missing", StatusFail:
		parts = append(parts, "✗ "+a.Detail)
	case StatusWarn:
		parts = append(parts, "⚠ "+a.Detail)
	default:
		if a.Detail != "" {
			parts = append(parts, a.Detail)
		}
	}
	return strings.Join(parts, " ")
}

func checkIcon(status string) string {
	switch status {
	case StatusOK:
		return "✓"
	case StatusWarn:
		return "⚠"
	default:
		return "✗"
	}
}

func writeHeader(w io.Writer, title string) {
	_, _ = fmt.Fprintln(w, title)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", len(title)))
}

func writeSection(w io.Writer, title string) {
	_, _ = fmt.Fprintf(w, "\n%s\n", title)
}

func trustVerdictLabel(v worktrust.Verdict) string {
	switch v {
	case worktrust.VerdictTrusted:
		return "Fiable"
	case worktrust.VerdictAcceptable:
		return "Acceptable"
	case worktrust.VerdictRisky:
		return "À surveiller"
	case worktrust.VerdictBlocked:
		return "Bloqué"
	default:
		return string(v)
	}
}

func trustVerdictRank(v worktrust.Verdict) int {
	switch v {
	case worktrust.VerdictBlocked:
		return 4
	case worktrust.VerdictRisky:
		return 3
	case worktrust.VerdictAcceptable:
		return 2
	case worktrust.VerdictTrusted:
		return 1
	default:
		return 0
	}
}
