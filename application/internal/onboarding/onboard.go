package onboarding

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding/detect"
)

// Onboard runs the full onboarding flow.
func Onboard(startDir string, opts Options, in io.Reader, out io.Writer) (Result, error) {
	repoRoot, err := gitRoot(startDir)
	if err != nil {
		return Result{}, err
	}

	if opts.CheckOnly {
		return Ready(startDir, opts, out)
	}

	var st State
	if opts.Resume {
		st, err = LoadState(repoRoot)
		if err != nil {
			return Result{}, err
		}
	} else {
		st = State{CurrentStep: StepWelcome}
	}

	st, err = RunWizard(repoRoot, st, opts, in, out)
	if err != nil {
		return Result{}, err
	}

	matches, validation := detect.DetectAll(repoRoot, firstNonEmpty(opts.Stack, st.Answers.Stack, "auto"))
	if len(validation) == 0 {
		validation = config.DefaultGoValidationCommandsForRepo(repoRoot)
	}
	if st.Answers.Stack == "" {
		st.Answers.Stack = detect.PrimaryStack(matches)
	}

	patch := ConfigPatch{
		ProjectName:     st.Answers.ProjectName,
		DefaultBranch:   st.Answers.DefaultBranch,
		BranchPrefix:    SlugFromName(st.Answers.ProjectName),
		DefaultAgent:    st.Answers.DefaultAgent,
		DefaultReviewer: st.Answers.DefaultReviewer,
		Validation:      validation,
	}

	cfgPath := config.ConfigPath(repoRoot)
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	if cfgErr != nil {
		cfg = config.NewTestConfig(filepath.Base(repoRoot))
	}

	merged, skipped := MergeConfig(cfg, patch, filepath.Base(repoRoot))
	var planned []PlannedChange
	planned = append(planned, PlannedChange{Path: config.DefaultConfigRel, Action: "update", Summary: "merge config"})

	docPlanned, docErr := BootstrapDocs(repoRoot, st.Answers, opts.ForceDocs, opts.DryRun)
	if docErr != nil {
		return Result{}, docErr
	}
	planned = append(planned, docPlanned...)

	var backupPath string
	if !opts.DryRun {
		backupPath, err = WriteConfig(repoRoot, cfgPath, merged, false)
		if err != nil {
			return Result{}, err
		}
		if err := SaveState(repoRoot, st); err != nil {
			return Result{}, err
		}
	}

	report, err := AssessReadiness(repoRoot, merged, opts.Strict)
	if err != nil {
		return Result{}, err
	}
	if !opts.DryRun {
		_ = PersistReport(repoRoot, report)
	}

	res := Result{
		Report:         report,
		PlannedChanges: planned,
		SkippedFields:  skipped,
		ConfigPath:     cfgPath,
		BackupPath:     backupPath,
	}
	if opts.DryRun {
		res.BackupPath = ""
	}

	if err := formatOnboardOutput(out, opts, res); err != nil {
		return res, err
	}
	return res, nil
}

// Ready assesses readiness without mutating config (check-only).
func Ready(startDir string, opts Options, out io.Writer) (Result, error) {
	repoRoot, err := gitRoot(startDir)
	if err != nil {
		return Result{}, err
	}
	cfgPath := config.ConfigPath(repoRoot)
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	if cfgErr != nil {
		cfg = nil
	}
	report, err := AssessReadiness(repoRoot, cfg, opts.Strict)
	if err != nil {
		return Result{}, err
	}
	var applied []AppliedAutofix
	if opts.Autofix && !report.Ready && len(ListAutofixOffers(report.Checks)) > 0 {
		applied, report, err = ApplyReadinessAutofixes(repoRoot)
		if err != nil {
			return Result{}, err
		}
	}
	if !opts.DryRun {
		_ = PersistReport(repoRoot, report)
	}
	res := Result{Report: report}
	if len(applied) > 0 {
		if err := formatAutofixPlain(out, applied, report); err != nil {
			return res, err
		}
	}
	if err := formatReadyOutput(out, opts, res); err != nil {
		return res, err
	}
	return res, nil
}

func formatAutofixPlain(out io.Writer, applied []AppliedAutofix, report Report) error {
	for _, fix := range applied {
		if fix.Description != "" {
			fmt.Fprintf(out, "Autofix: %s\n", fix.Description)
		}
		for _, line := range fix.AddedLines {
			fmt.Fprintf(out, "  + %s\n", line)
		}
	}
	status := "NOT READY"
	if report.Ready {
		status = "READY"
	}
	fmt.Fprintf(out, "Après corrections: %s (%d/100)\n\n", status, report.Score)
	return nil
}

func formatOnboardOutput(out io.Writer, opts Options, res Result) error {
	if opts.JSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}
	if opts.DryRun {
		fmt.Fprintln(out, "Dry-run — changements prévus :")
		for _, p := range res.PlannedChanges {
			fmt.Fprintf(out, "  [%s] %s %s\n", p.Action, p.Path, p.Summary)
		}
		return nil
	}
	if opts.Plain || opts.CI {
		return formatReadyPlain(out, res.Report)
	}
	fmt.Fprintf(out, "Onboarding terminé — ready=%v score=%d\n", res.Report.Ready, res.Report.Score)
	if len(res.SkippedFields) > 0 {
		fmt.Fprintf(out, "Champs conservés: %s\n", strings.Join(res.SkippedFields, ", "))
	}
	return nil
}

func formatReadyOutput(out io.Writer, opts Options, res Result) error {
	if opts.JSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(res.Report)
	}
	return formatReadyPlain(out, res.Report)
}

func formatReadyPlain(out io.Writer, report Report) error {
	status := "NOT READY"
	if report.Ready {
		status = "READY"
	}
	fmt.Fprintf(out, "Readiness: %s (score %d/100)\n\n", status, report.Score)
	for _, c := range report.Checks {
		line := fmt.Sprintf("[%s] %s", c.Status, c.ID)
		if c.Message != "" {
			line += ": " + c.Message
		}
		fmt.Fprintln(out, line)
	}
	if len(report.NextActions) > 0 {
		fmt.Fprintln(out, "\nNext actions:")
		for _, a := range report.NextActions {
			fmt.Fprintf(out, "  - %s → %s\n", a.Title, a.CLI)
		}
	}
	if offers := ListAutofixOffers(report.Checks); len(offers) > 0 && !report.Ready {
		fmt.Fprintln(out, "\nCorrections auto disponibles (asa ready --autofix):")
		for _, o := range offers {
			fmt.Fprintf(out, "  - %s\n", o.Title)
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
