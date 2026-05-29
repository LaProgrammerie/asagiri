package cost

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/routing"
)

// BuildOpts tunes estimation without importing pipeline.
type BuildOpts struct {
	RunID               string
	MaxOutputTokens     int
	DefaultOutputTokens int
}

// BuildEstimate projects tokens, cost, and duration for a plan (specv3 §3.3).
func BuildEstimate(ctx context.Context, plan intent.ExecutionPlan, inv investigation.InvestigationResult, pack contextopt.ContextPack, cfg *config.Config, dm DurationModel, opts BuildOpts) (ExecutionEstimate, error) {
	if cfg == nil {
		return ExecutionEstimate{}, fmt.Errorf("estimate: config nil")
	}
	if dm == nil {
		dm = DefaultDurationModel{Cfg: cfg}
	}
	packText := contextopt.PackText(pack)
	invText := stringsJoinLines(inv.CandidateFiles) + "\n" + stringsJoinLines(inv.GrepHits)
	if opts.DefaultOutputTokens <= 0 {
		opts.DefaultOutputTokens = 4096
	}
	if opts.MaxOutputTokens > 0 {
		opts.DefaultOutputTokens = opts.MaxOutputTokens
	}

	est := ExecutionEstimate{
		RunID:        opts.RunID,
		Feature:      plan.Intent.Feature,
		TaskID:       plan.Intent.TaskID,
		Confidence:   0.65,
		BudgetStatus: BudgetOK,
	}

	llmSteps := 0
	for _, s := range plan.Steps {
		if stepUsesCloudModel(s.Command) {
			llmSteps++
		}
	}
	outSplit := opts.DefaultOutputTokens
	if llmSteps > 1 {
		outSplit = opts.DefaultOutputTokens / llmSteps
		if outSplit < 512 {
			outSplit = 512
		}
	}

	agent := cfg.Work.DefaultAgent
	reviewer := cfg.Work.DefaultReviewer
	enricher := cfg.Work.DefaultEnricher
	totalIn, totalOut := 0, 0
	var totalCost Money
	cur := cfg.Pricing.Currency
	if cur == "" {
		cur = cfg.Budgets.DefaultCurrency
	}
	totalDur := time.Duration(0)

	defDM, _ := dm.(DefaultDurationModel)

	for _, s := range plan.Steps {
		step := estimatedFromPlanStep(s, cfg, agent, reviewer, enricher, packText, invText, outSplit)
		if stepUsesCloudModel(s.Command) {
			rd := routing.Route(cfg, s.Command, false, false, true)
			if step.Reason == "" {
				step.Reason = rd.Reason
			} else {
				step.Reason = step.Reason + "; routing=" + rd.Reason
			}
			if rd.Agent != "" && step.Agent == agent {
				step.Agent = rd.Agent
			}
		}
		hist := RunHistory{}
		if defDM.Reader != nil {
			hist.RecentDurations = defDM.Reader.DurationsFor(step.Name, step.Model, 8)
		}
		ps := PlannedStep{
			Name:        step.Name,
			Agent:       step.Agent,
			Model:       step.Model,
			Local:       step.Local,
			InputTokens: step.InputTokens,
			TaskType:    string(plan.Intent.Action),
		}
		step.EstimatedDuration = dm.Estimate(ctx, ps, hist)

		if !step.Local && step.Model != "" && step.Model != "unconfigured" {
			mny, err := CostFromPricing(cfg, step.Model, step.InputTokens, step.OutputTokens)
			if err != nil {
				est.Warnings = append(est.Warnings, err.Error())
				mny = Money{Currency: cur}
			}
			step.EstimatedCost = mny
			totalCost.Cents += mny.Cents
			if totalCost.Currency == "" {
				totalCost.Currency = mny.Currency
			}
		}
		totalIn += step.InputTokens
		totalOut += step.OutputTokens
		totalDur += step.EstimatedDuration
		est.PlannedSteps = append(est.PlannedSteps, step)
	}

	est.TotalInputTokens = totalIn
	est.TotalOutputTokens = totalOut
	est.TotalTokens = totalIn + totalOut
	est.EstimatedCost = totalCost
	if est.EstimatedCost.Currency == "" {
		est.EstimatedCost.Currency = cur
	}
	est.EstimatedDuration = totalDur
	est.Risk = "medium"
	return est, nil
}

func stringsJoinLines(a []string) string {
	return strings.Join(a, "\n")
}

func stepUsesCloudModel(cmd string) bool {
	switch cmd {
	case "spec", "enrich", "dev", "review":
		return true
	default:
		return false
	}
}

func estimatedFromPlanStep(step intent.PlanStep, cfg *config.Config, agent, reviewer, enricher, packText, invText string, outTok int) EstimatedStep {
	name := step.Command
	local := !stepUsesCloudModel(step.Command)
	ag := agent
	mdl := ""
	inTok := 0
	reason := step.Reason
	switch step.Command {
	case "spec":
		ag = intentAgentName(step.Args, "kiro")
		mdl = cfg.AgentModel(ag)
	case "enrich":
		ag = intentAgentName(step.Args, enricher)
		mdl = cfg.AgentModel(ag)
		local = isLocalAgent(cfg, ag)
	case "dev":
		ag = intentAgentName(step.Args, agent)
		mdl = cfg.AgentModel(ag)
		local = isLocalAgent(cfg, ag)
	case "review":
		ag = intentAgentName(step.Args, reviewer)
		mdl = cfg.AgentModel(ag)
		local = isLocalAgent(cfg, ag)
	default:
		local = true
		inTok = 500
		outTok = 0
	}
	if inTok == 0 && !local {
		inTok = CountTokens(packText, mdl, ContentDefault, cfg.TokenEst) +
			CountTokens(invText, mdl, ContentDefault, cfg.TokenEst)
	}
	out := outTok
	if local {
		out = 0
	}
	if mdl == "" && !local {
		mdl = "unconfigured"
	}
	return EstimatedStep{
		Name:         name,
		Agent:        ag,
		Model:        mdl,
		Local:        local,
		InputTokens:  inTok,
		OutputTokens: out,
		Reason:       reason,
	}
}

func intentAgentName(args []string, def string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--agent" {
			return args[i+1]
		}
	}
	return def
}

func isLocalAgent(cfg *config.Config, name string) bool {
	if cfg == nil {
		return false
	}
	a, ok := cfg.Agents[name]
	if ok && strings.Contains(strings.ToLower(a.Endpoint), "localhost") {
		return true
	}
	if p, ok := cfg.Models[name]; ok {
		return strings.Contains(strings.ToLower(p.Class), "local")
	}
	return name == "ollama"
}
