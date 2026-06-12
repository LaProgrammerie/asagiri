package onboarding

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentsync"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding/detect"
)

// RunWizard executes onboarding steps; non-interactive mode uses defaults only.
func RunWizard(repoRoot string, st State, opts Options, in io.Reader, out io.Writer) (State, error) {
	st = MergeAnswers(st, opts, repoRoot)

	if opts.Step != "" {
		step, err := ParseStep(opts.Step)
		if err != nil {
			return st, err
		}
		st.CurrentStep = step
	}

	startIdx := StepIndex(st.CurrentStep)
	if startIdx < 0 {
		startIdx = 0
		st.CurrentStep = stepOrder[0]
	}

	interactive := !opts.NonInteractive && !opts.Yes && isTerminal(in)
	for i := startIdx; i < len(stepOrder); i++ {
		st.CurrentStep = stepOrder[i]
		if err := runStep(repoRoot, &st, opts, interactive, in, out); err != nil {
			return st, err
		}
		if opts.DryRun && st.CurrentStep == StepReview {
			break
		}
	}
	return st, nil
}

func runStep(repoRoot string, st *State, opts Options, interactive bool, in io.Reader, out io.Writer) error {
	switch st.CurrentStep {
	case StepWelcome:
		if interactive {
			_, _ = fmt.Fprintln(out, "Bienvenue — Asagiri onboarding prépare config, docs et validation.")
		}
	case StepProject:
		if interactive && st.Answers.ProjectName == "" {
			st.Answers.ProjectName = promptString(in, out, "Nom du projet", filepathBase(repoRoot))
		}
		if interactive && st.Answers.DefaultBranch == "" {
			st.Answers.DefaultBranch = promptString(in, out, "Branche par défaut", "main")
		}
	case StepStack:
		if st.Answers.Stack == "" {
			matches, cmds := detect.DetectAll(repoRoot, opts.Stack)
			st.Answers.Stack = detect.PrimaryStack(matches)
			if st.Answers.Stack == "" && len(cmds) > 0 {
				st.Answers.Stack = "mixed"
			}
		}
		if interactive {
			_, _ = fmt.Fprintf(out, "Stack détectée: %s\n", st.Answers.Stack)
		}
	case StepProviders:
		if strings.TrimSpace(st.Answers.EnabledProviders) == "" {
			st.Answers.EnabledProviders = formatEnabledProvidersCSV(DefaultEnabledProviders())
		}
		if interactive {
			_, _ = fmt.Fprintln(out, "Étape 1 — Select providers (runtimes externes, adapter = provider.type)")
			for _, p := range ProviderCatalog() {
				_, _ = fmt.Fprintf(out, "  • %s (%s)\n", p.Label, p.ID)
			}
			st.Answers.EnabledProviders = promptString(in, out, "Providers activés (csv)", st.Answers.EnabledProviders)
		}
	case StepAgents:
		if interactive {
			_, _ = fmt.Fprintln(out, "Étape 2 — Create logical agents (profils dans config.agents, référencés par work.*)")
			_, _ = fmt.Fprintf(out, "  Exemples: %s\n", strings.Join(DefaultLogicalAgentNames(), ", "))
			if st.Answers.DefaultSpecAgent == "" {
				st.Answers.DefaultSpecAgent = promptString(in, out, "Agent spec (work.default_spec_agent)", config.DefaultAgentSpec)
			}
			if st.Answers.DefaultEnricher == "" {
				st.Answers.DefaultEnricher = promptString(in, out, "Agent enrich (work.default_enricher)", config.DefaultAgentEnrich)
			}
			if st.Answers.DefaultAgent == "" {
				st.Answers.DefaultAgent = promptString(in, out, "Agent dev (work.default_agent)", config.DefaultAgentDev)
			}
			if st.Answers.DefaultReviewer == "" {
				st.Answers.DefaultReviewer = promptString(in, out, "Agent review (work.default_reviewer)", config.DefaultAgentReviewer)
			}
		}
		if !opts.DryRun {
			report, err := agentsync.Sync(repoRoot, agentsync.Options{Write: true})
			if err != nil {
				return err
			}
			for _, item := range report.Items {
				if item.Action == agentsync.ActionCreate {
					_, _ = fmt.Fprintf(out, "  + AgentSpec %s (%s)\n", item.ID, item.Path)
				}
			}
		} else if interactive {
			report, err := agentsync.Sync(repoRoot, agentsync.Options{})
			if err != nil {
				return err
			}
			if report.Hint != "" {
				_, _ = fmt.Fprintf(out, "  → %s\n", report.Hint)
			}
		}
	case StepSources, StepDocs:
		// defaults from MergeAnswers
	case StepFeature:
		if st.Answers.FeatureSlug == "" {
			st.Answers.FeatureSlug = SlugFromName(st.Answers.ProjectName) + "-mvp"
		}
		if interactive {
			st.Answers.FeatureSlug = promptString(in, out, "Slug première feature Kiro", st.Answers.FeatureSlug)
		}
	case StepReview:
		if interactive {
			_, _ = fmt.Fprintln(out, "Récap — écriture config et docs au step validate.")
		}
	case StepValidate:
		return nil
	}
	return nil
}

func promptString(in io.Reader, out io.Writer, label, def string) string {
	if !isTerminal(in) {
		return def
	}
	_, _ = fmt.Fprintf(out, "%s [%s]: ", label, def)
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return def
	}
	v := strings.TrimSpace(scanner.Text())
	if v == "" {
		return def
	}
	return v
}

func isTerminal(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func filepathBase(repoRoot string) string {
	parts := strings.Split(strings.TrimRight(repoRoot, string(os.PathSeparator)), string(os.PathSeparator))
	if len(parts) == 0 {
		return "project"
	}
	return parts[len(parts)-1]
}
