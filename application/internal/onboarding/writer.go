package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"gopkg.in/yaml.v3"
)

// MergeConfig applies a patch while preserving user-edited non-template values.
func MergeConfig(existing *config.Config, patch ConfigPatch, repoDirName string) (*config.Config, []string) {
	if existing == nil {
		existing = config.NewTestConfig(repoDirName)
	}
	out := *existing
	skipped := make([]string, 0, 8)

	if patch.ProjectName != "" && config.IsTemplateDefaultProjectName(out.Project.Name) {
		out.Project.Name = patch.ProjectName
	} else if patch.ProjectName != "" && out.Project.Name != patch.ProjectName {
		skipped = append(skipped, "project.name")
	}
	if patch.DefaultBranch != "" && strings.TrimSpace(out.Project.DefaultBranch) == "" {
		out.Project.DefaultBranch = patch.DefaultBranch
	} else if patch.DefaultBranch != "" && out.Project.DefaultBranch == "main" && patch.DefaultBranch != "main" {
		out.Project.DefaultBranch = patch.DefaultBranch
	}
	if patch.BranchPrefix != "" && config.IsTemplateDefaultBranchPrefix(out.Worktrees.BranchPrefix) {
		out.Worktrees.BranchPrefix = patch.BranchPrefix
	} else if patch.BranchPrefix != "" && out.Worktrees.BranchPrefix != patch.BranchPrefix {
		skipped = append(skipped, "worktrees.branch_prefix")
	}
	if patch.DefaultAgent != "" {
		out.Work.DefaultAgent = patch.DefaultAgent
	}
	if patch.DefaultReviewer != "" {
		out.Work.DefaultReviewer = patch.DefaultReviewer
	}
	if len(patch.Validation) > 0 {
		if config.IsTemplateDefaultValidationCommands(out.Validation.Commands) {
			out.Validation.Commands = dedupeValidation(patch.Validation)
		} else {
			out.Validation.Commands = mergeValidationCommands(out.Validation.Commands, patch.Validation)
		}
	}
	return &out, skipped
}

func dedupeValidation(cmds []config.ValidationCommand) []config.ValidationCommand {
	seen := map[string]bool{}
	out := make([]config.ValidationCommand, 0, len(cmds))
	for _, c := range cmds {
		key := strings.TrimSpace(c.Command)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

func mergeValidationCommands(existing, proposed []config.ValidationCommand) []config.ValidationCommand {
	index := map[string]config.ValidationCommand{}
	for _, c := range existing {
		index[strings.TrimSpace(c.Command)] = c
	}
	for _, c := range proposed {
		key := strings.TrimSpace(c.Command)
		if key == "" {
			continue
		}
		if _, ok := index[key]; !ok {
			index[key] = c
		}
	}
	out := make([]config.ValidationCommand, 0, len(index))
	for _, c := range index {
		out = append(out, c)
	}
	return out
}

// WriteConfig serializes cfg to configPath with optional backup.
func WriteConfig(repoRoot, configPath string, cfg *config.Config, dryRun bool) (backupPath string, err error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("sérialiser config: %w", err)
	}
	if dryRun {
		return "", nil
	}
	backupDir := filepath.Join(repoRoot, backupsRel)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}
	if _, statErr := os.Stat(configPath); statErr == nil {
		backupPath = filepath.Join(backupDir, "config.yaml."+fmt.Sprintf("%d", time.Now().Unix()))
		src, err := os.ReadFile(configPath)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(backupPath, src, 0o644); err != nil {
			return "", fmt.Errorf("backup config: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return "", fmt.Errorf("écrire config: %w", err)
	}
	return backupPath, nil
}

// SlugFromName converts a project name to a branch prefix slug.
func SlugFromName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "project"
	}
	return s
}
