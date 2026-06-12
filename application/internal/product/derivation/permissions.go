package derivation

import "strings"

func derivePermissionsRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		for _, action := range flow.StepActions {
			if action.Sensitive {
				out = append(out, "sensitive_action:"+flow.ID+"."+action.Action)
				out = append(out, "human_review_required:"+flow.ID+"."+action.Action)
			}
			if strings.Contains(strings.ToLower(action.Action), "invite") {
				out = append(out, "rate_limit:"+flow.ID+"."+action.Action)
				out = append(out, "owner_only:"+flow.ID+"."+action.Action)
			}
		}
	}
	return out
}

func derivePermissionsMatrix(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		for _, action := range flow.StepActions {
			out = append(out, "owner:"+action.Action)
			if !action.Sensitive {
				out = append(out, "member:"+action.Action)
			}
		}
	}
	return out
}
