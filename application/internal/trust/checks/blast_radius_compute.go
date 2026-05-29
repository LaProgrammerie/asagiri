package checks

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

func computeBlastRadius(flow product.Flow, bundle analysis.Bundle, bundleErr error) BlastRadiusSummary {
	summary := BlastRadiusSummary{
		MigrationRisk:      "low",
		PublicContractRisk: "low",
	}
	if bundleErr != nil {
		summary.MigrationRisk = "unknown"
		summary.PublicContractRisk = "medium"
		return summary
	}

	flowGraph := bundle.Graphs["flow"]
	summary.FlowsImpacted = len(flowGraph.Nodes)
	if summary.FlowsImpacted == 0 && flow.ID != "" {
		summary.FlowsImpacted = 1
	}

	apiGraph := bundle.Graphs["api"]
	for _, n := range apiGraph.Nodes {
		if n.Kind == "route" || strings.Contains(n.ID, "route:") {
			summary.CriticalAPIs++
		}
	}
	for _, step := range flow.Steps {
		ref := strings.TrimSpace(step.ContractRef)
		if ref != "" && !strings.HasPrefix(ref, "TODO:") && !graphHasRoute(apiGraph, ref) {
			summary.PublicContractRisk = "high"
		}
	}

	depGraph := bundle.Graphs["dependency"]
	outDegree := make(map[string]int)
	for _, e := range depGraph.Edges {
		outDegree[e.From]++
	}
	for _, n := range depGraph.Nodes {
		if outDegree[n.ID] >= 2 {
			summary.SharedModules++
		}
	}
	if summary.SharedModules == 0 && len(depGraph.Nodes) > 0 {
		summary.SharedModules = len(depGraph.Nodes)
	}

	if len(unresolvedContractRefs(flow)) > 0 {
		summary.PublicContractRisk = "high"
	} else if summary.CriticalAPIs > 0 && summary.PublicContractRisk != "high" {
		summary.PublicContractRisk = "medium"
	}

	switch {
	case summary.SharedModules >= 5 || flow.Business.Criticality == "high" && summary.FlowsImpacted > 2:
		summary.MigrationRisk = "high"
	case summary.SharedModules >= 2 || flow.Business.Criticality == "high":
		summary.MigrationRisk = "medium"
	default:
		summary.MigrationRisk = "low"
	}
	return summary
}
