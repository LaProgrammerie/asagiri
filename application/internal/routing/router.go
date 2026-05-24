package routing

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Decision is model/agent routing outcome.
type Decision struct {
	Agent      string
	Model      string
	Local      bool
	StepClass  string
	Reason     string
}

// Route picks agent/model for a step class (specv3 §11).
func Route(cfg *config.Config, stepClass string, preferLocal, noCloud, allowCloud bool) Decision {
	stepClass = strings.ToLower(stepClass)
	d := Decision{StepClass: stepClass, Agent: "cursor", Model: ""}
	if cfg != nil {
		d.Agent = cfg.Work.DefaultAgent
		d.Model = cfg.AgentModel(d.Agent)
	}

	strategy := cfg.Routing.Strategies["cost_aware"]
	if cfg.Routing.DefaultStrategy != "" {
		if s, ok := cfg.Routing.Strategies[cfg.Routing.DefaultStrategy]; ok {
			strategy = s
		}
	}

	if preferLocal || contains(strategy.PreferLocalFor, stepClass) {
		d.Local = true
		d.Agent = cfg.Work.DefaultEnricher
		if d.Agent == "" {
			d.Agent = "ollama"
		}
		d.Model = cfg.AgentModel(d.Agent)
		d.Reason = "prefer_local"
		return d
	}
	if noCloud && !allowCloud {
		d.Local = true
		d.Reason = "no_cloud"
		return d
	}
	if contains(strategy.UseCloudHeavyFor, stepClass) {
		d.Agent = "claude"
		d.Model = cfg.AgentModel(d.Agent)
		d.Reason = "cloud_heavy"
		return d
	}
	if contains(strategy.UseCloudFastFor, stepClass) {
		d.Agent = cfg.Work.DefaultAgent
		d.Model = cfg.AgentModel(d.Agent)
		d.Reason = "cloud_fast"
		return d
	}
	d.Reason = "default"
	return d
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if strings.EqualFold(x, item) {
			return true
		}
	}
	return false
}
