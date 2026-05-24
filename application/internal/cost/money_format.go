package cost

import "fmt"

// String renders Money as major currency units (2 decimals).
func (m Money) String() string {
	if m.Currency == "" {
		return fmt.Sprintf("%.2f", float64(m.Cents)/100)
	}
	return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Cents)/100)
}
