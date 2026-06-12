package app

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) prototypeInputActive() bool {
	if m.showHelp || m.showPalette || m.confirmation != nil {
		return false
	}
	return m.router.Current() == ScreenPrototype
}

func (m *model) updatePrototypeKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.prototypeInputActive() {
		return m, nil
	}
	product := strings.TrimSpace(m.snapshot.Prototype.Product)
	if product == "" {
		product = strings.TrimSpace(m.prototypeProduct)
	}
	switch v.String() {
	case "1":
		return m, m.dispatchCommand(bus.PrototypeCreateCommand{
			Intent:  "workspace onboarding prototype",
			Product: product,
		}, `asa prototype create "<intent>"`)
	case "2":
		if product == "" {
			m.lastCommandResult = "prototype product required for flows extract"
			return m, nil
		}
		return m, m.dispatchCommand(bus.FlowsExtractCommand{Product: product}, "asa flows extract "+product)
	case "3":
		if product == "" {
			m.lastCommandResult = "prototype product required for contracts extract"
			return m, nil
		}
		return m, m.dispatchCommand(bus.ContractsExtractCommand{Product: product}, "asa contracts extract "+product)
	case "4":
		if product == "" {
			m.lastCommandResult = "prototype product required for spec generate"
			return m, nil
		}
		return m, m.dispatchCommand(bus.SpecGenerateFromProductCommand{Product: product}, "asa spec generate-from-product "+product)
	}
	return m, nil
}

func (m model) dispatchPrototypePipelineAction(cli string) tea.Cmd {
	product := strings.TrimSpace(m.snapshot.Prototype.Product)
	if product == "" {
		product = strings.TrimSpace(m.prototypeProduct)
	}
	cli = strings.TrimSpace(cli)
	switch {
	case strings.HasPrefix(cli, "asa prototype create"):
		return m.dispatchCommand(bus.PrototypeCreateCommand{
			Intent:  "workspace onboarding prototype",
			Product: product,
		}, cli)
	case strings.HasPrefix(cli, "asa flows extract"):
		if product == "" {
			return nil
		}
		return m.dispatchCommand(bus.FlowsExtractCommand{Product: product}, cli)
	case strings.HasPrefix(cli, "asa contracts extract"):
		if product == "" {
			return nil
		}
		return m.dispatchCommand(bus.ContractsExtractCommand{Product: product}, cli)
	case strings.HasPrefix(cli, "asa spec generate-from-product"):
		if product == "" {
			return nil
		}
		return m.dispatchCommand(bus.SpecGenerateFromProductCommand{Product: product}, cli)
	default:
		return nil
	}
}
