package derivation

import "strings"

func deriveAnalyticsRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		for _, metric := range flow.Metrics {
			out = append(out, "metric:"+metric)
			out = append(out, "dashboard:"+metric)
		}
		for _, step := range flow.StepActions {
			out = append(out, "event:"+flow.ID+"."+step.Action)
		}
	}
	return out
}

func deriveMetricsCoverage(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		if len(flow.Metrics) == 0 {
			out = append(out, "missing_metrics:"+flow.ID)
			continue
		}
		for _, metric := range flow.Metrics {
			out = append(out, "coupled:"+flow.ID+":"+strings.TrimSpace(metric))
		}
	}
	return out
}
