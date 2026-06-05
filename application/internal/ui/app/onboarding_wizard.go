package app

import (
	"fmt"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	onboardscreen "github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	tea "github.com/charmbracelet/bubbletea"
)

type onboardingWizardMsg struct {
	result bus.OnboardingWizardResult
	err    error
}

type onboardingAdvanceMsg struct {
	form    onbdomain.Form
	err     error
	message string
}

type onboardingApplyMsg struct {
	readiness bus.ReadinessResult
	err       error
	message   string
}

type onboardingAutofixMsg struct {
	readiness bus.ReadinessResult
	err       error
	message   string
}

func (m model) onboardingInputActive() bool {
	if m.showHelp || m.showPalette || m.confirmation != nil {
		return false
	}
	return m.router.Current() == ScreenOnboarding
}

func (m model) loadOnboardingWizardCmd() tea.Cmd {
	return func() tea.Msg {
		if m.queryBus == nil {
			return onboardingWizardMsg{err: fmt.Errorf("query bus unavailable")}
		}
		res, err := m.queryBus.Query(m.ctx, bus.GetOnboardingWizardQuery{})
		if err != nil {
			return onboardingWizardMsg{err: err}
		}
		typed, ok := res.(bus.OnboardingWizardResult)
		if !ok {
			return onboardingWizardMsg{err: fmt.Errorf("invalid wizard query result")}
		}
		return onboardingWizardMsg{result: typed}
	}
}

func (m *model) ensureOnboardingWizard() {
	if m.onboardingWizard.Initialized {
		return
	}
	if m.queryBus == nil {
		return
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetOnboardingWizardQuery{})
	if err != nil {
		m.lastError = err.Error()
		return
	}
	typed, ok := res.(bus.OnboardingWizardResult)
	if !ok {
		return
	}
	m.onboardingWizard = wizardFromResult(typed, m.cfg.Mouse)
}

func wizardFromResult(res bus.OnboardingWizardResult, mouse bool) onboardscreen.Model {
	step := onbdomain.WizardStep(res.CurrentStep)
	form := onbdomain.FormFromMaps(step, res.Fields, res.Advanced)
	form.ValidationPreview = res.ValidationPreview
	form.DetectedStacks = res.DetectedStacks
	form.SkippedFields = res.SkippedFields
	form.Errors = res.Errors
	form.HasAsagiriConfig = res.HasAsagiriConfig
	return onboardscreen.NewModelFromForm(form, mouse)
}

func (m *model) updateOnboardingKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.ensureOnboardingWizard()
	next, cmd := m.onboardingWizard.Update(v)
	m.onboardingWizard = next
	return m, cmd
}

func (m *model) updateOnboardingMouse(v tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.onboardingInputActive() || !m.cfg.Mouse {
		return m, nil
	}
	m.ensureOnboardingWizard()
	next, cmd := m.onboardingWizard.Update(v)
	m.onboardingWizard = next
	return m, cmd
}

func (m model) handleOnboardingFooter(msg onboardscreen.OnboardingFooterMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case onboardscreen.FooterPrev:
		return m, m.advanceOnboardingCmd("prev", msg.Form, false)
	case onboardscreen.FooterNext:
		return m, m.advanceOnboardingCmd("next", msg.Form, true)
	case onboardscreen.FooterAdvanced:
		m.onboardingWizard.ShowAdvanced = !m.onboardingWizard.ShowAdvanced
		m.onboardingWizard.RefreshFieldRows()
		return m, nil
	case onboardscreen.FooterApply:
		return m, m.applyOnboardingCmd(msg.Form)
	default:
		return m, nil
	}
}

func (m model) advanceOnboardingCmd(direction string, form onbdomain.Form, validate bool) tea.Cmd {
	return func() tea.Msg {
		if m.commandBus == nil {
			return onboardingAdvanceMsg{err: fmt.Errorf("command bus unavailable")}
		}
		if validate {
			if m.queryBus != nil {
				res, err := m.queryBus.Query(m.ctx, bus.ValidateOnboardingStepQuery{
					Step:     string(form.Step),
					Fields:   form.FieldsMap(),
					Advanced: form.AdvancedMap(),
				})
				if err != nil {
					return onboardingAdvanceMsg{form: form, err: err}
				}
				if typed, ok := res.(bus.ValidateOnboardingStepResult); ok && !typed.Valid {
					form.Errors = typed.Errors
					return onboardingAdvanceMsg{form: form, err: fmt.Errorf("validation")}
				}
			}
		}
		_, err := m.commandBus.Dispatch(m.ctx, bus.AdvanceOnboardingStepCommand{
			Direction: direction,
			Fields:    form.FieldsMap(),
			Advanced:  form.AdvancedMap(),
		})
		if err != nil {
			return onboardingAdvanceMsg{form: form, err: err}
		}
		next, advErr := onbdomain.AdvanceTUIStep(form, direction, validate)
		if advErr != nil {
			return onboardingAdvanceMsg{form: next, err: advErr}
		}
		return onboardingAdvanceMsg{form: next}
	}
}

func (m model) applyOnboardingCmd(form onbdomain.Form) tea.Cmd {
	return func() tea.Msg {
		if m.commandBus == nil {
			return onboardingApplyMsg{err: fmt.Errorf("command bus unavailable")}
		}
		res, err := m.commandBus.Dispatch(m.ctx, bus.ApplyOnboardingConfigCommand{
			Yes:      true,
			Fields:   form.FieldsMap(),
			Advanced: form.AdvancedMap(),
		})
		if err != nil {
			return onboardingApplyMsg{err: err}
		}
		readiness := bus.ReadinessResult{}
		if m.queryBus != nil {
			if q, qErr := m.queryBus.Query(m.ctx, bus.GetReadinessQuery{}); qErr == nil {
				if typed, ok := q.(bus.ReadinessResult); ok {
					readiness = typed
				}
			}
		}
		return onboardingApplyMsg{readiness: readiness, message: res.Message}
	}
}

func (m model) handleOnboardingWizardLoaded(msg onboardingWizardMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.lastError = msg.err.Error()
		return m, nil
	}
	m.onboardingWizard = wizardFromResult(msg.result, m.cfg.Mouse)
	return m, nil
}

func (m model) handleOnboardingAdvance(msg onboardingAdvanceMsg) (tea.Model, tea.Cmd) {
	m.onboardingWizard.SyncForm(msg.form)
	if msg.err != nil {
		m.onboardingWizard.Message = "Corrigez les champs invalides"
		m.lastError = msg.err.Error()
		return m, nil
	}
	m.onboardingWizard.Message = ""
	m.lastError = ""
	m.lastCommandResult = msg.message
	return m, m.snapshotQueryCmd()
}

func (m model) applyAutofixCmd() tea.Cmd {
	return func() tea.Msg {
		if m.commandBus == nil {
			return onboardingAutofixMsg{err: fmt.Errorf("command bus unavailable")}
		}
		res, err := m.commandBus.Dispatch(m.ctx, bus.ApplyReadinessAutofixCommand{})
		if err != nil {
			return onboardingAutofixMsg{err: err}
		}
		readiness := bus.ReadinessResult{}
		if m.queryBus != nil {
			if q, qErr := m.queryBus.Query(m.ctx, bus.GetReadinessQuery{}); qErr == nil {
				if typed, ok := q.(bus.ReadinessResult); ok {
					readiness = typed
				}
			}
		}
		return onboardingAutofixMsg{readiness: readiness, message: res.Message}
	}
}

func (m model) handleOnboardingApply(msg onboardingApplyMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.lastError = msg.err.Error()
		m.onboardingWizard.Message = "Échec apply"
		return m, nil
	}
	m.onboardingWizard.Applied = true
	m.onboardingWizard.Readiness = msg.readiness
	m.onboardingWizard.Message = msg.message
	m.lastError = ""
	m.lastCommandResult = msg.message
	// CK-4.4: once the repo is ready, hand off to Mission Control through the
	// shared shell. Otherwise keep the ready summary so autofixes can be applied.
	if msg.readiness.Ready {
		m.wizardMode = false
		m.navigateTo(ScreenMission, "asa")
	}
	return m, m.snapshotQueryCmd()
}

func (m model) handleOnboardingAutofixRequest() (tea.Model, tea.Cmd) {
	return m, m.applyAutofixCmd()
}

func (m model) handleOnboardingAutofix(msg onboardingAutofixMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.lastError = msg.err.Error()
		m.onboardingWizard.Message = "Échec corrections auto"
		return m, nil
	}
	m.onboardingWizard.Readiness = msg.readiness
	m.onboardingWizard.Message = msg.message
	m.lastError = ""
	m.lastCommandResult = msg.message
	return m, m.snapshotQueryCmd()
}
