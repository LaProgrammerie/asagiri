package onboarding

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// AssessReadiness builds a readiness report for the repository.
func AssessReadiness(repoRoot string, cfg *config.Config, strict bool) (Report, error) {
	checks := RunDoctorChecks(repoRoot, cfg, DoctorOpts{Full: true, SkipExec: true})
	report := Report{Score: initialScore, Checks: checks}
	for _, c := range checks {
		switch c.Status {
		case StatusFail:
			report.Score -= scoreFail
		case StatusWarn:
			report.Score -= scoreWarn
		}
	}
	if report.Score < 0 {
		report.Score = 0
	}
	report.Ready = true
	for _, c := range checks {
		if c.Status == StatusFail {
			report.Ready = false
			break
		}
		if strict && c.Status == StatusWarn {
			report.Ready = false
			break
		}
	}
	report.NextActions = deriveNextActions(checks)
	return report, nil
}

func deriveNextActions(checks []Check) []Action {
	var actions []Action
	for _, c := range checks {
		if c.Status == StatusOK {
			continue
		}
		title := c.Message
		if title == "" {
			title = c.ID
		}
		cli := c.FixCLI
		if cli == "" {
			cli = "asa onboard --step " + stepForCheck(c.ID)
		}
		actions = append(actions, Action{Title: title, CLI: cli})
	}
	return actions
}

func stepForCheck(id string) string {
	switch {
	case strings.HasPrefix(id, "docs."):
		return "docs"
	case strings.HasPrefix(id, "spec."):
		return "feature"
	case strings.HasPrefix(id, "agents."):
		return "agents"
	case strings.HasPrefix(id, "validation."):
		return "stack"
	default:
		return "validate"
	}
}

// PersistReport writes report.json under .asagiri/onboarding/.
func PersistReport(repoRoot string, report Report) error {
	dir := filepath.Join(repoRoot, dirRel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(repoRoot, reportRel)
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// DoctorOpts controls extended doctor checks.
type DoctorOpts struct {
	Full     bool
	Strict   bool
	SkipExec bool
}

// RunDoctorChecks returns onboarding-aware checks (used by doctor --full and ready).
func RunDoctorChecks(repoRoot string, cfg *config.Config, opts DoctorOpts) []Check {
	var checks []Check

	if cfg != nil {
		if err := cfg.Validate(repoRoot); err != nil {
			checks = append(checks, Check{ID: "config.valid", Status: StatusFail, Message: err.Error(), FixCLI: "asa onboard"})
		} else {
			checks = append(checks, Check{ID: "config.valid", Status: StatusOK})
		}
	} else {
		checks = append(checks, Check{ID: "config.valid", Status: StatusFail, Message: "config.yaml introuvable ou invalide", FixCLI: "asa init"})
	}

	if !opts.Full {
		return checks
	}

	checks = append(checks, checkGitignore(repoRoot)...)
	checks = append(checks, checkAgents(cfg)...)
	checks = append(checks, checkDocsPlaceholders(repoRoot)...)
	checks = append(checks, checkKiroSpec(repoRoot, cfg)...)
	checks = append(checks, checkMacOSAsaConflict()...)

	if cfg != nil && !opts.SkipExec {
		checks = append(checks, checkValidationCommands(cfg)...)
	}

	return checks
}

func checkGitignore(repoRoot string) []Check {
	path := filepath.Join(repoRoot, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return []Check{{
			ID:      "gitignore.asagiri",
			Status:  StatusFail,
			Message: ".gitignore manquant",
			FixCLI:  "asa onboard --autofix",
		}}
	}
	content := string(data)
	var out []Check
	if gitignoreHasRequiredLine(content, ".asagiri/state.sqlite") {
		out = append(out, Check{ID: "gitignore._asagiri_state.sqlite", Status: StatusOK})
	} else {
		out = append(out, Check{
			ID:      "gitignore._asagiri_state.sqlite",
			Status:  StatusFail,
			Message: ".gitignore doit contenir .asagiri/state.sqlite",
			FixCLI:  "asa onboard --autofix",
		})
	}
	if gitignoreHasWorktrees(content) {
		out = append(out, Check{ID: "gitignore.worktrees_", Status: StatusOK})
	} else {
		out = append(out, Check{
			ID:      "gitignore.worktrees_",
			Status:  StatusFail,
			Message: ".gitignore doit contenir .asagiri/worktrees/ (ou worktrees/)",
			FixCLI:  "asa onboard --autofix",
		})
	}
	return out
}

func checkAgents(cfg *config.Config) []Check {
	if cfg == nil {
		return nil
	}
	agentName := cfg.Work.DefaultAgent
	if agentName == "" {
		agentName = "cursor"
	}
	agent, ok := cfg.Agents[agentName]
	if !ok {
		return []Check{{
			ID:      "agents." + agentName,
			Status:  StatusWarn,
			Message: fmt.Sprintf("agent %q absent de config.agents", agentName),
			FixCLI:  "asa onboard --step agents",
		}}
	}
	cmd := strings.TrimSpace(agent.Command)
	if cmd == "" {
		return []Check{{
			ID:      "agents." + agentName,
			Status:  StatusWarn,
			Message: fmt.Sprintf("agents.%s.command vide", agentName),
			FixCLI:  "asa onboard --step agents",
		}}
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return []Check{{
			ID:      "agents." + agentName,
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s not in PATH", cmd),
			FixCLI:  "asa onboard --step agents",
		}}
	}
	return []Check{{ID: "agents." + agentName, Status: StatusOK}}
}

func checkDocsPlaceholders(repoRoot string) []Check {
	path := filepath.Join(repoRoot, "docs", "ai", "01-product.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return []Check{{
			ID:      "docs.product",
			Status:  StatusWarn,
			Message: "docs/ai/01-product.md manquant",
			FixCLI:  "asa onboard --step docs",
		}}
	}
	if isPlaceholderContent(string(data)) {
		return []Check{{
			ID:      "docs.product",
			Status:  StatusWarn,
			Message: "docs/ai/01-product.md encore placeholder",
			FixCLI:  "asa onboard --step docs",
		}}
	}
	return []Check{{ID: "docs.product", Status: StatusOK}}
}

func checkKiroSpec(repoRoot string, cfg *config.Config) []Check {
	kiroPath := ".kiro/specs"
	if cfg != nil && cfg.Specs.KiroPath != "" {
		kiroPath = cfg.Specs.KiroPath
	}
	abs := filepath.Join(repoRoot, kiroPath)
	entries, err := os.ReadDir(abs)
	if err != nil {
		return []Check{{
			ID:      "spec.kiro",
			Status:  StatusFail,
			Message: "aucune feature sous " + kiroPath + "/",
			FixCLI:  "asa onboard --step feature",
		}}
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}
	if count == 0 {
		return []Check{{
			ID:      "spec.kiro",
			Status:  StatusFail,
			Message: "aucune feature sous " + kiroPath + "/",
			FixCLI:  "asa onboard --step feature",
		}}
	}
	return []Check{{ID: "spec.kiro", Status: StatusOK}}
}

func checkMacOSAsaConflict() []Check {
	if runtime.GOOS != "darwin" {
		return nil
	}
	path := "/usr/bin/asa"
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if info.Mode()&0o111 != 0 {
		return []Check{{
			ID:      "system.asa_conflict",
			Status:  StatusWarn,
			Message: "binaire système /usr/bin/asa détecté — peut masquer le CLI Asagiri",
			FixCLI:  "which asa",
		}}
	}
	return []Check{{ID: "system.asa_conflict", Status: StatusOK}}
}

func checkValidationCommands(cfg *config.Config) []Check {
	if cfg == nil || len(cfg.Validation.Commands) == 0 {
		return []Check{{
			ID:      "validation.commands",
			Status:  StatusWarn,
			Message: "validation.commands vide",
			FixCLI:  "asa onboard --step stack",
		}}
	}
	return []Check{{ID: "validation.commands", Status: StatusOK}}
}

func isPlaceholderContent(content string) bool {
	lower := strings.ToLower(content)
	markers := []string{
		"template",
		"placeholder",
		"après fork",
		"my-project",
		"remplace ce paragraphe",
		"à compléter",
	}
	hits := 0
	for _, m := range markers {
		if strings.Contains(lower, m) {
			hits++
		}
	}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	return hits >= 1 && nonEmpty < 15
}
