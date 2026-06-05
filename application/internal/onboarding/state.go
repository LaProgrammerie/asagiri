package onboarding

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// WizardStep identifies one onboarding step.
type WizardStep string

const (
	StepWelcome  WizardStep = "welcome"
	StepProject  WizardStep = "project"
	StepStack    WizardStep = "stack"
	StepAgents   WizardStep = "agents"
	StepSources  WizardStep = "sources"
	StepDocs     WizardStep = "docs"
	StepFeature  WizardStep = "feature"
	StepReview   WizardStep = "review"
	StepValidate WizardStep = "validate"
)

var stepOrder = []WizardStep{
	StepWelcome,
	StepProject,
	StepStack,
	StepAgents,
	StepSources,
	StepDocs,
	StepFeature,
	StepReview,
	StepValidate,
}

// State persists wizard progress under .asagiri/onboarding/state.json.
type State struct {
	CurrentStep WizardStep `json:"current_step"`
	Answers     Answers    `json:"answers"`
	Completed   []string   `json:"completed,omitempty"`
}

func statePath(repoRoot string) string {
	return filepath.Join(repoRoot, stateRel)
}

// LoadState reads persisted wizard state; missing file returns zero state at welcome.
func LoadState(repoRoot string) (State, error) {
	path := statePath(repoRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{CurrentStep: StepWelcome}, nil
		}
		return State{}, fmt.Errorf("lire state onboarding: %w", err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}, fmt.Errorf("parser state onboarding: %w", err)
	}
	if st.CurrentStep == "" {
		st.CurrentStep = StepWelcome
	}
	return st, nil
}

// SaveState writes wizard state (creates parent dirs).
func SaveState(repoRoot string, st State) error {
	dir := filepath.Join(repoRoot, dirRel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("créer %s: %w", dirRel, err)
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	path := statePath(repoRoot)
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("écrire state onboarding: %w", err)
	}
	return nil
}

// StepIndex returns the index of a step or -1.
func StepIndex(step WizardStep) int {
	for i, s := range stepOrder {
		if s == step {
			return i
		}
	}
	return -1
}

// ParseStep validates a step name from CLI.
func ParseStep(name string) (WizardStep, error) {
	switch WizardStep(name) {
	case StepWelcome, StepProject, StepStack, StepAgents, StepSources, StepDocs, StepFeature, StepReview, StepValidate:
		return WizardStep(name), nil
	default:
		return "", fmt.Errorf("step inconnu %q (project|stack|agents|docs|feature|validate)", name)
	}
}

// AdvanceStep moves to the next step in order.
func AdvanceStep(st State) State {
	idx := StepIndex(st.CurrentStep)
	if idx < 0 || idx >= len(stepOrder)-1 {
		return st
	}
	st.CurrentStep = stepOrder[idx+1]
	return st
}

// MergeAnswers overlays non-empty opts onto state answers.
func MergeAnswers(st State, opts Options, repoRoot string) State {
	a := st.Answers
	if opts.ProjectName != "" {
		a.ProjectName = opts.ProjectName
	} else if a.ProjectName == "" {
		a.ProjectName = filepath.Base(repoRoot)
	}
	if opts.DefaultBranch != "" {
		a.DefaultBranch = opts.DefaultBranch
	} else if a.DefaultBranch == "" {
		a.DefaultBranch = "main"
	}
	if opts.Tagline != "" {
		a.Tagline = opts.Tagline
	}
	if opts.Stack != "" && opts.Stack != "auto" {
		a.Stack = opts.Stack
	}
	if opts.DefaultSpecAgent != "" {
		a.DefaultSpecAgent = opts.DefaultSpecAgent
	} else if a.DefaultSpecAgent == "" {
		a.DefaultSpecAgent = config.DefaultAgentSpec
	}
	if opts.DefaultAgent != "" {
		a.DefaultAgent = opts.DefaultAgent
	} else if a.DefaultAgent == "" {
		a.DefaultAgent = config.DefaultAgentDev
	}
	if opts.DefaultReviewer != "" {
		a.DefaultReviewer = opts.DefaultReviewer
	} else if a.DefaultReviewer == "" {
		a.DefaultReviewer = config.DefaultAgentReviewer
	}
	if opts.DefaultEnricher != "" {
		a.DefaultEnricher = opts.DefaultEnricher
	} else if a.DefaultEnricher == "" {
		a.DefaultEnricher = config.DefaultAgentEnrich
	}
	if opts.FeatureSlug != "" {
		a.FeatureSlug = opts.FeatureSlug
	}
	if opts.ProductOneLiner != "" {
		a.ProductOneLiner = opts.ProductOneLiner
	}
	if opts.ProductUsers != "" {
		a.ProductUsers = opts.ProductUsers
	}
	st.Answers = a
	return st
}
