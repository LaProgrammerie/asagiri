package product

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/product/derivation"
)

// ValidateProjectionCouplingStrict enforces a hard-fail policy for
// metrics/analytics/contracts coupling.
func ValidateProjectionCouplingStrict(p derivation.Projection) error {
	missing := make([]string, 0)
	coupledMetrics := make(map[string]struct{})
	for _, item := range p.MetricsCoverage {
		switch {
		case strings.HasPrefix(item, "missing_metrics:"):
			missing = append(missing, strings.TrimPrefix(item, "missing_metrics:"))
		case strings.HasPrefix(item, "coupled:"):
			parts := strings.Split(item, ":")
			if len(parts) >= 3 {
				metric := strings.TrimSpace(parts[len(parts)-1])
				if metric != "" {
					coupledMetrics[metric] = struct{}{}
				}
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("strict coupling failed: missing flow metrics for %s", strings.Join(missing, ", "))
	}

	analyticsSet := make(map[string]struct{}, len(p.Analytics))
	for _, item := range p.Analytics {
		analyticsSet[item] = struct{}{}
	}
	observabilitySet := make(map[string]struct{}, len(p.Observability))
	for _, item := range p.Observability {
		observabilitySet[item] = struct{}{}
	}

	for metric := range coupledMetrics {
		if _, ok := analyticsSet["metric:"+metric]; !ok {
			return fmt.Errorf("strict coupling failed: analytics metric missing for %s", metric)
		}
		if _, ok := analyticsSet["dashboard:"+metric]; !ok {
			return fmt.Errorf("strict coupling failed: analytics dashboard missing for %s", metric)
		}
		if _, ok := observabilitySet["metric:"+metric]; !ok {
			return fmt.Errorf("strict coupling failed: observability metric missing for %s", metric)
		}
	}
	return nil
}

