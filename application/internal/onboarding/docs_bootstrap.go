package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const substantiveLineThreshold = 20

// BootstrapDocs fills safe template docs and Kiro spec skeleton.
func BootstrapDocs(repoRoot string, answers Answers, force bool, dryRun bool) ([]PlannedChange, error) {
	var planned []PlannedChange
	slug := answers.FeatureSlug
	if slug == "" {
		slug = SlugFromName(answers.ProjectName) + "-mvp"
	}
	name := answers.ProjectName
	if name == "" {
		name = filepath.Base(repoRoot)
	}
	oneLiner := answers.ProductOneLiner
	if oneLiner == "" {
		oneLiner = answers.Tagline
	}
	if oneLiner == "" {
		oneLiner = name + " — produit en cours de définition"
	}
	users := answers.ProductUsers
	if users == "" {
		users = "Développeurs et équipes techniques"
	}

	writes := []struct {
		rel     string
		content string
	}{
		{
			rel: "docs/ai/01-product.md",
			content: fmt.Sprintf(`# Produit

## Problème résolu

%s

## Utilisateurs / contexte

%s

## État

| Tranche | Statut |
|---------|--------|
| bootstrap | En cours |
`, oneLiner, users),
		},
		{
			rel:     "docs/ai/active/current-spec.md",
			content: fmt.Sprintf("# Spec active\n\nFeature Kiro : **%s**\n\nVoir `.kiro/specs/%s/`.\n", slug, slug),
		},
		{
			rel: "docs/ai/active/handoff.md",
			content: `# Handoff — execution

> Contrat d'exécution pour l'implémentation.

## Objectif immédiat

Bootstrap environnement et première feature **` + slug + `**.

## Plan

1. Compléter requirements / design Kiro
2. Lancer ` + "`asa work`" + `
`,
		},
		{
			rel: filepath.Join(".kiro", "specs", slug, "requirements.md"),
			content: fmt.Sprintf(`# %s — Requirements

## Objectif

%s

## Critères d'acceptation

- [ ] Environnement prêt (`+"`asa ready`"+`)
- [ ] Première tâche implémentée
`, slug, oneLiner),
		},
		{
			rel: filepath.Join(".kiro", "specs", slug, "design.md"),
			content: fmt.Sprintf(`# %s — Design

## Vue d'ensemble

À compléter après validation produit.

## Stack

Détectée par `+"`asa onboard`"+`.
`, slug),
		},
		{
			rel: filepath.Join(".kiro", "specs", slug, "tasks.md"),
			content: `# Tasks

- [ ] 1. Bootstrap env — ` + "`asa ready`" + ` OK
`,
		},
	}

	for _, w := range writes {
		abs := filepath.Join(repoRoot, w.rel)
		action := "create"
		if _, err := os.Stat(abs); err == nil {
			if !force && isSubstantiveFile(abs) {
				planned = append(planned, PlannedChange{Path: w.rel, Action: "skip", Summary: "contenu substantiel conservé"})
				continue
			}
			action = "update"
		}
		planned = append(planned, PlannedChange{Path: w.rel, Action: action})
		if dryRun {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return planned, err
		}
		if err := os.WriteFile(abs, []byte(w.content), 0o644); err != nil {
			return planned, fmt.Errorf("écrire %s: %w", w.rel, err)
		}
	}

	if err := patchAGENTS(repoRoot, name, oneLiner, force, dryRun); err != nil {
		return planned, err
	}
	planned = append(planned, PlannedChange{Path: "AGENTS.md", Action: "patch"})
	return planned, nil
}

func isSubstantiveFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	if nonEmpty >= substantiveLineThreshold {
		return true
	}
	return !isPlaceholderContent(string(data))
}

func patchAGENTS(repoRoot, name, oneLiner string, force bool, dryRun bool) error {
	path := filepath.Join(repoRoot, "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	content := string(data)
	if !force && !strings.Contains(content, "Template") && !strings.Contains(content, "Après fork") {
		return nil
	}
	replacement := fmt.Sprintf("**%s** : %s", name, oneLiner)
	updated := strings.Replace(content,
		"**Template** : squelette **Go** (`application/`, `go.mod`) + **Docker Compose** local (`infrastructure/docker/`) + couche [AI Engineering](https://github.com/LaProgrammerie/ai-engineering-framework) (`docs/ai/`, `.kiro/`, `.cursor/`). Orchestration via **Makefile** — pas Castor, pas Yoimachi.",
		replacement, 1)
	if updated == content {
		updated = strings.Replace(content,
			"**Après fork / copie :** remplace ce paragraphe par **une phrase** sur le produit ou service réel.",
			replacement, 1)
	}
	if updated == content {
		return nil
	}
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}
