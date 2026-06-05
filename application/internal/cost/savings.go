package cost

import (
	"fmt"
	"math"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

// SavingsReport is the cost intelligence output for one time window.
type SavingsReport struct {
	// Always populated
	LocalInputTokens  int64
	LocalOutputTokens int64
	CloudInputTokens  int64
	CloudOutputTokens int64
	ActualCostCents   int64
	Currency          string

	// Populated only when PremiumReferenceModel is configured
	PremiumReferenceModel string
	PremiumEquivCents     int64
	SavingsCents          int64
	SavingsRate           float64 // 0..1; 0 when no reference model
}

// HasPremiumBaseline reports whether savings figures are meaningful.
func (r SavingsReport) HasPremiumBaseline() bool {
	return r.PremiumReferenceModel != ""
}

// TotalInputTokens returns local + cloud input tokens.
func (r SavingsReport) TotalInputTokens() int64 { return r.LocalInputTokens + r.CloudInputTokens }

// TotalOutputTokens returns local + cloud output tokens.
func (r SavingsReport) TotalOutputTokens() int64 { return r.LocalOutputTokens + r.CloudOutputTokens }

// LocalPct returns the fraction of total tokens that ran locally (0..100).
func (r SavingsReport) LocalPct() float64 {
	total := r.TotalInputTokens() + r.TotalOutputTokens()
	if total == 0 {
		return 0
	}
	local := r.LocalInputTokens + r.LocalOutputTokens
	return float64(local) / float64(total) * 100
}

// ComputeSavings builds a SavingsReport from token data and config.
// When cfg.Pricing.PremiumReferenceModel is empty, savings fields are zero.
// No baseline is invented: the function never returns non-zero savings without
// an explicit reference model configured by the user.
func ComputeSavings(tokens telemetry.StepTokenTotals, actualCents int64, cfg *config.Config) SavingsReport {
	cur := "EUR"
	if cfg != nil && cfg.Pricing.Currency != "" {
		cur = cfg.Pricing.Currency
	}
	r := SavingsReport{
		LocalInputTokens:  tokens.LocalInputTokens,
		LocalOutputTokens: tokens.LocalOutputTokens,
		CloudInputTokens:  tokens.CloudInputTokens,
		CloudOutputTokens: tokens.CloudOutputTokens,
		ActualCostCents:   actualCents,
		Currency:          cur,
	}
	if cfg == nil || strings.TrimSpace(cfg.Pricing.PremiumReferenceModel) == "" {
		return r // no baseline → no savings figures
	}
	refModel := strings.TrimSpace(cfg.Pricing.PremiumReferenceModel)
	p, ok := cfg.Pricing.Models[refModel]
	if !ok {
		return r // reference model not in pricing table → no savings figures
	}
	// Premium equivalent = what all tokens would have cost at premium prices.
	totalIn := r.TotalInputTokens()
	totalOut := r.TotalOutputTokens()
	premiumMajor := float64(totalIn)/1_000_000*p.InputPer1MTokens +
		float64(totalOut)/1_000_000*p.OutputPer1MTokens
	premiumCents := int64(math.Round(premiumMajor * 100))

	r.PremiumReferenceModel = refModel
	r.PremiumEquivCents = premiumCents
	r.SavingsCents = premiumCents - actualCents
	if premiumCents > 0 {
		r.SavingsRate = float64(r.SavingsCents) / float64(premiumCents)
	}
	return r
}

// FormatCents formats a cents value as a currency string (e.g. "€4.90").
func FormatCents(cents int64, currency string) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	sym := currencySymbol(currency)
	return fmt.Sprintf("%s%s%d.%02d", sign, sym, cents/100, cents%100)
}

func currencySymbol(currency string) string {
	switch strings.ToUpper(currency) {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	case "GBP":
		return "£"
	default:
		return currency + " "
	}
}
