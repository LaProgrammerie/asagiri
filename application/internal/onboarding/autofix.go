package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const gitignoreAutofixID = "gitignore.asagiri"

var requiredGitignoreLines = []string{
	".asagiri/state.sqlite",
	".asagiri/worktrees/",
}

// AutofixOffer describes one safe automatic correction the wizard can apply.
type AutofixOffer struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Lines       []string `json:"lines,omitempty"`
	CheckIDs    []string `json:"check_ids,omitempty"`
}

// AppliedAutofix records one applied correction.
type AppliedAutofix struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	AddedLines  []string `json:"added_lines,omitempty"`
	Description string   `json:"description,omitempty"`
}

// ListAutofixOffers returns safe automatic fixes for failed checks.
func ListAutofixOffers(checks []Check) []AutofixOffer {
	if offer := gitignoreAutofixOffer(checks); offer != nil {
		return []AutofixOffer{*offer}
	}
	return nil
}

func gitignoreAutofixOffer(checks []Check) *AutofixOffer {
	var failed []Check
	for _, c := range checks {
		if c.Status != StatusFail {
			continue
		}
		switch c.ID {
		case "gitignore.asagiri", "gitignore._asagiri_state.sqlite", "gitignore.worktrees_":
			failed = append(failed, c)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	ids := make([]string, 0, len(failed))
	for _, c := range failed {
		ids = append(ids, c.ID)
	}
	return &AutofixOffer{
		ID:          gitignoreAutofixID,
		Title:       ".gitignore — entrées Asagiri",
		Description: "Ajoute les chemins locaux Asagiri ignorés par git",
		Lines:       append([]string(nil), requiredGitignoreLines...),
		CheckIDs:    ids,
	}
}

// ApplyReadinessAutofixes applies all safe automatic fixes and re-assesses readiness.
func ApplyReadinessAutofixes(repoRoot string) ([]AppliedAutofix, Report, error) {
	cfgPath := config.ConfigPath(repoRoot)
	cfg, err := config.Load(cfgPath, repoRoot)
	if err != nil {
		cfg = nil
	}
	checks := RunDoctorChecks(repoRoot, cfg, DoctorOpts{Full: true, SkipExec: true})
	offers := ListAutofixOffers(checks)
	if len(offers) == 0 {
		report, assessErr := AssessReadiness(repoRoot, cfg, false)
		return nil, report, assessErr
	}

	var applied []AppliedAutofix
	for _, offer := range offers {
		switch offer.ID {
		case gitignoreAutofixID:
			fix, fixErr := applyGitignoreAutofix(repoRoot)
			if fixErr != nil {
				return applied, Report{}, fixErr
			}
			if fix != nil {
				applied = append(applied, *fix)
			}
		}
	}

	report, err := AssessReadiness(repoRoot, cfg, false)
	if err != nil {
		return applied, Report{}, err
	}
	_ = PersistReport(repoRoot, report)
	return applied, report, nil
}

func applyGitignoreAutofix(repoRoot string) (*AppliedAutofix, error) {
	path := filepath.Join(repoRoot, ".gitignore")
	content, err := os.ReadFile(path)
	missing := false
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		missing = true
	}
	text := string(content)
	var added []string
	for _, line := range requiredGitignoreLines {
		if gitignoreHasRequiredLine(text, line) {
			continue
		}
		if len(text) > 0 && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if missing && len(text) == 0 {
			text = "# Asagiri (ajouté par le wizard)\n"
		}
		text += line + "\n"
		added = append(added, line)
	}
	if len(added) == 0 {
		return nil, nil
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return nil, err
	}
	return &AppliedAutofix{
		ID:          gitignoreAutofixID,
		Path:        ".gitignore",
		AddedLines:  added,
		Description: fmt.Sprintf("Ajouté %d entrée(s) dans .gitignore", len(added)),
	}, nil
}

func gitignoreHasRequiredLine(content, required string) bool {
	if required == ".asagiri/worktrees/" {
		return gitignoreHasWorktrees(content)
	}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == required {
			return true
		}
	}
	return false
}

func gitignoreHasWorktrees(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch line {
		case "worktrees/", ".asagiri/worktrees/", ".asagiri/worktrees":
			return true
		}
	}
	return false
}
