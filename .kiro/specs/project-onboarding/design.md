# project-onboarding — Design

## Package layout

```
application/internal/onboarding/
  wizard.go           # step machine, prompts (huh or plain stdin)
  state.go            # persist/resume .asagiri/onboarding/state.json
  writer.go           # config merge + backup
  readiness.go        # score, checks, next actions
  docs_bootstrap.go   # safe template fills
  detect/
    detect.go         # registry
    go.go
    castor.go
    node.go
  detect/detect_test.go
  fixtures/           # mini repos for tests
```

## Wizard steps

| Step ID | Collects | Writes |
|---------|----------|--------|
| `welcome` | — | — |
| `project` | name, branch, tagline | `project.*` |
| `stack` | confirm detected / override | `validation.commands` |
| `agents` | default_agent, reviewer | `work.*`, `agents.*` |
| `sources` | local paths | `sources.local`, `specs.*` |
| `docs` | product one-liner, users | `docs/ai/01-product.md`, AGENTS.md |
| `feature` | first feature slug | `.kiro/specs/<slug>/` |
| `review` | confirm | all pending |
| `validate` | run doctor + ready | `report.json` |

Non-interactive: `--yes --non-interactive` uses detection + defaults; skips prompts.

## Stack detector contract

```go
type StackMatch struct {
    ID          string   // "go", "castor", "node"
    Confidence  float64
    Signals     []string // files that matched
}

type Detector interface {
    ID() string
    Detect(repoRoot string) (StackMatch, error)
    ValidationCommands(repoRoot string) []config.ValidationCommand
}
```

Registry runs all detectors; merges validation commands deduplicated by command string.

## Config merge rules

1. Load existing config if present.
2. Apply patches only for "template defaults" (helper `config.IsTemplateDefault(field)`).
3. User-edited values preserved; wizard logs skipped fields.
4. Write backup to `.asagiri/onboarding/backups/config.yaml.<unix>`.

## Readiness model

```go
type Check struct {
    ID      string
    Status  string // ok | warn | fail
    Message string
    FixCLI  string // optional
}

type Report struct {
    Ready       bool
    Score       int
    Checks      []Check
    NextActions []Action
}
```

Scoring: start 100, `-20` fail, `-5` warn (configurable). `Ready = no fail`.

## Doctor integration

Extend `bootstrap.Doctor` or add `onboarding.RunDoctorChecks(repoRoot, cfg, opts)` called from:
- `asa doctor --full`
- final onboard step
- `asa ready`

## CLI wiring

- `newOnboardCmd()` in `cli/onboard_cmd.go`
- `newReadyCmd()` alias
- Extend `newDoctorCmd()` with `--full`

## UI (lot 2)

- `ScreenOnboarding` in `ui/screens/onboarding/`
- Query: `GetReadinessQuery`, `GetOnboardingProgressQuery`
- Command: `AdvanceOnboardingStepCommand`, `RunOnboardingValidateCommand`
- Mission Control: if `!report.Ready`, render `components.ReadinessBanner`

## Testing strategy

- Table tests per detector with fixture dirs under `internal/onboarding/fixtures/`
- Integration: temp git repo → Init → Onboard(yes) → Ready == true
- Golden files for `ready --plain`

## Dependencies

- Existing: `config`, `bootstrap`, `env`
- Optional interactive: `charmbracelet/huh` (already in go.mod for UI)
- No new external deps required for lot 1

## Security

- Do not write secrets into config.yaml
- `secret_path_denylist` unchanged
- Doc bootstrap must not read `.env`
