package runtime

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HooksConfig lists event-triggered commands (spec-my-A §24.9).
type HooksConfig struct {
	Hooks struct {
		On map[string][]HookAction `yaml:"on"`
	} `yaml:"hooks"`
}

// HookAction is one hook entry.
type HookAction struct {
	Run string `yaml:"run"`
}

// LoadHooks reads `.asagiri/runtime/hooks.yaml` if present.
func LoadHooks(repoRoot string) (HooksConfig, error) {
	var cfg HooksConfig
	path := filepath.Join(repoRoot, DefaultRelDir, "hooks.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// DispatchHooks enqueues configured hooks for an event type.
func (s *Store) DispatchHooks(eventType string, dryRun bool) []string {
	cfg, err := LoadHooks(s.repoRoot)
	if err != nil || cfg.Hooks.On == nil {
		return nil
	}
	actions, ok := cfg.Hooks.On[eventType]
	if !ok {
		return nil
	}
	var ran []string
	for _, a := range actions {
		if a.Run == "" {
			continue
		}
		ran = append(ran, a.Run)
		if dryRun {
			continue
		}
		_ = s.EnqueueHook(eventType, a.Run)
		_, _ = s.EmitEvent("hook.scheduled", "hooks", "", "", map[string]any{"command": a.Run, "on": eventType})
	}
	return ran
}
