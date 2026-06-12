package agentsync

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
)

// Sync plans or applies embedded AgentSpec templates into .asagiri/agents.
func Sync(repoRoot string, opts Options) (Report, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return Report{}, fmt.Errorf("agentsync: repo_root requis")
	}

	templates, err := agentspec.ListEmbeddedTemplates()
	if err != nil {
		return Report{}, err
	}
	if id := strings.TrimSpace(opts.AgentID); id != "" {
		filtered := make([]agentspec.EmbeddedTemplate, 0, 1)
		for _, tpl := range templates {
			if tpl.ID == id {
				filtered = append(filtered, tpl)
				break
			}
		}
		if len(filtered) == 0 {
			return Report{}, fmt.Errorf("agentsync: template embarqué %q introuvable", id)
		}
		templates = filtered
	}

	mode := "check"
	if opts.Write {
		mode = "write"
	}

	report := Report{
		ReportVersion: ReportVersion,
		Mode:          mode,
		RegistryDir:   agentspec.RegistryDir,
		Wrote:         false,
		Items:         make([]Item, 0, len(templates)),
	}

	registryDir := filepath.Join(repoRoot, agentspec.RegistryDir)
	if opts.Write {
		if err := os.MkdirAll(registryDir, 0o755); err != nil {
			return Report{}, fmt.Errorf("agentsync: création %s: %w", registryDir, err)
		}
	}

	for _, tpl := range templates {
		item, wrote, err := syncOne(registryDir, tpl, opts)
		if err != nil {
			return report, err
		}
		if wrote {
			report.Wrote = true
		}
		report.Items = append(report.Items, item)
	}

	if hasAction(report.Items, ActionConflict) {
		report.Hint = "Conflits détectés — relancer avec --force pour écraser, ou résoudre manuellement"
	} else if hasPending(report.Items) && !opts.Write {
		report.Hint = "asa agents sync --write"
	} else if report.Wrote {
		report.Hint = "Registry synchronisé"
	}
	return report, nil
}

func syncOne(registryDir string, tpl agentspec.EmbeddedTemplate, opts Options) (Item, bool, error) {
	targetPath := filepath.Join(registryDir, tpl.ID+".yaml")
	relPath := filepath.Join(agentspec.RegistryDir, tpl.ID+".yaml")

	templateSpec, err := agentspec.Parse(tpl.Data, "embedded")
	if err != nil {
		return Item{}, false, err
	}
	templateHash := templateSpec.ContentHash

	data, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return Item{}, false, fmt.Errorf("agentsync: lecture %s: %w", targetPath, readErr)
		}
		item := Item{
			ID:      tpl.ID,
			Action:  ActionCreate,
			Path:    relPath,
			Message: "fichier absent — création depuis template embarqué",
		}
		if !opts.Write {
			return item, false, nil
		}
		if err := os.WriteFile(targetPath, tpl.Data, 0o644); err != nil {
			return Item{}, false, fmt.Errorf("agentsync: écriture %s: %w", targetPath, err)
		}
		return item, true, nil
	}

	diskSpec, err := agentspec.Parse(data, targetPath)
	if err != nil {
		return Item{}, false, fmt.Errorf("agentsync: spec disque %s: %w", targetPath, err)
	}

	if diskSpec.ContentHash == templateHash {
		return Item{
			ID:      tpl.ID,
			Action:  ActionSkip,
			Path:    relPath,
			Message: "identique au template embarqué",
		}, false, nil
	}

	item := Item{
		ID:      tpl.ID,
		Action:  ActionConflict,
		Path:    relPath,
		Message: "fichier modifié — différent du template embarqué",
		Diff:    summarizeDiff(templateHash, diskSpec.ContentHash),
	}
	if !opts.Force {
		if opts.Write {
			item.Message += " (non écrasé sans --force)"
		}
		return item, false, nil
	}

	item.Action = ActionUpdate
	item.Message = "écrasé depuis template embarqué (--force)"
	if !opts.Write {
		return item, false, nil
	}
	if err := os.WriteFile(targetPath, tpl.Data, 0o644); err != nil {
		return Item{}, false, fmt.Errorf("agentsync: écriture %s: %w", targetPath, err)
	}
	return item, true, nil
}

func summarizeDiff(templateHash, diskHash string) string {
	return fmt.Sprintf("template_hash=%s disk_hash=%s", truncateHash(templateHash), truncateHash(diskHash))
}

func truncateHash(h string) string {
	h = strings.TrimSpace(h)
	if len(h) <= 12 {
		return h
	}
	return h[:12] + "…"
}

func hasAction(items []Item, action string) bool {
	for _, it := range items {
		if it.Action == action {
			return true
		}
	}
	return false
}

func hasPending(items []Item) bool {
	for _, it := range items {
		switch it.Action {
		case ActionCreate, ActionUpdate, ActionConflict:
			return true
		}
	}
	return false
}

// HasBlockingConflicts reports unresolved conflicts after sync.
func HasBlockingConflicts(report Report) bool {
	return hasAction(report.Items, ActionConflict)
}

// FormatJSON writes report as JSON to w.
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatText renders a human-readable sync report.
func FormatText(w io.Writer, report Report) error {
	_, err := fmt.Fprintf(w, "Asagiri Agents Sync (%s)\n", report.Mode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Registry: %s\n\n", report.RegistryDir)
	if err != nil {
		return err
	}
	for _, it := range report.Items {
		if err := formatItem(w, it); err != nil {
			return err
		}
	}
	if h := strings.TrimSpace(report.Hint); h != "" {
		_, err = fmt.Fprintf(w, "\n→ %s\n", h)
	}
	return err
}

func formatItem(w io.Writer, it Item) error {
	_, err := fmt.Fprintf(w, "• %s  %s  %s\n", it.ID, it.Action, it.Path)
	if err != nil {
		return err
	}
	if msg := strings.TrimSpace(it.Message); msg != "" {
		_, err = fmt.Fprintf(w, "    %s\n", msg)
		if err != nil {
			return err
		}
	}
	if d := strings.TrimSpace(it.Diff); d != "" {
		_, err = fmt.Fprintf(w, "    %s\n", d)
	}
	return err
}
