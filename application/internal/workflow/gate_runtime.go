package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"gopkg.in/yaml.v3"
)

type governanceAgentHook = gateAgentHook
type planGateAgentHook = gateAgentHook

type gateAgentHook func(ctx context.Context, prompt string) (stdout string, err error)

func (s *Service) gateDryRunResult(gateID, gateType, scope, dryRunNote string, evidence []gates.EvidenceRef) gates.Result {
	return gates.Result{
		GateID:     gateID,
		GateType:   gateType,
		Scope:      scope,
		Status:     gates.VerdictPass,
		Confidence: 1,
		Notes:      []string{dryRunNote},
		Evidence:   evidence,
		DryRun:     true,
	}
}

func (s *Service) executeGateAgent(
	ctx context.Context,
	agentName, feature, taskID, workDir, prompt string,
	hook gateAgentHook,
) (stdout string, err error) {
	if hook != nil {
		stdout, err = hook(ctx, prompt)
		if err != nil {
			return stdout, fmt.Errorf("gate agent: %w", err)
		}
		return stdout, nil
	}
	a, err := s.ensureAgent(agentName)
	if err != nil {
		return "", err
	}
	req := agent.RunRequest{
		Feature:    feature,
		Prompt:     prompt,
		WorkingDir: workDir,
	}
	if taskID != "" {
		req.TaskID = taskID
	}
	res, runErr := a.Run(ctx, req)
	stdout = res.Stdout
	if stdout == "" {
		stdout = res.Stderr
	}
	if runErr != nil {
		return stdout, fmt.Errorf("gate agent run: %w", runErr)
	}
	return stdout, nil
}

func (s *Service) persistGateLogs(scopeID, scopeKind, gateName, feature, agentName, blockKey, logTitle, agentStdout string, r gates.Result) error {
	at := time.Now().UTC().Format(time.RFC3339)
	doc := gates.NewLogDocument(scopeID, scopeKind, gateName, feature, agentName, r, at)
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := s.writeGateScopeLog(scopeID, gateName, ".json", string(body)+"\n"); err != nil {
		return err
	}
	return s.writeGateMarkdownLog(scopeID, gateName, logTitle, blockKey, agentStdout, r)
}

func (s *Service) writeGateScopeLog(scopeID, gateName, ext, body string) error {
	logDir := filepath.Join(s.repoRoot, ".asagiri", "logs", scopeID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create gate log dir: %w", err)
	}
	logPath := filepath.Join(logDir, gateName+ext)
	if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write gate log: %w", err)
	}
	return nil
}

func (s *Service) writeGateMarkdownLog(scopeID, gateName, title, blockKey, agentStdout string, r gates.Result) error {
	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n\n## Agent stdout\n\n")
	if strings.TrimSpace(agentStdout) == "" {
		sb.WriteString("(empty)\n\n")
	} else {
		sb.WriteString(agentStdout)
		if !strings.HasSuffix(agentStdout, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("## Verdict\n\n")
	verdictYAML, err := yaml.Marshal(map[string]any{
		blockKey: map[string]any{
			"status":      string(r.Status),
			"confidence":  r.Confidence,
			"notes":       r.Notes,
			"findings":    r.Findings,
			"evidence":    r.Evidence,
			"dry_run":     r.DryRun,
			"parse_error": r.ParseError,
		},
	})
	if err != nil {
		_, _ = fmt.Fprintf(&sb, "status: %s\n", r.Status)
	} else {
		sb.Write(verdictYAML)
	}
	return s.writeGateScopeLog(scopeID, gateName, ".log", sb.String())
}

func gateOutcomeError(label string, r gates.Result, warnAdvisory bool) error {
	switch r.Status {
	case gates.VerdictPass:
		return nil
	case gates.VerdictWarn:
		if warnAdvisory {
			return nil
		}
		return fmt.Errorf("%s warn (non-advisory): %s", label, gates.FormatFailure(r))
	case gates.VerdictFail:
		return fmt.Errorf("%s failed: %s", label, gates.FormatFailure(r))
	default:
		return fmt.Errorf("%s unknown status %q", label, r.Status)
	}
}
