package app

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

func (m *model) setFocusContext(kind bus.FocusContextKind, subject, detail string) {
	m.state.SetFocusContext(bus.FocusContext{
		Kind:    kind,
		Subject: subject,
		Detail:  detail,
		Screen:  m.router.Current(),
	})
}

func (m model) explainContext() bus.ExplainContext {
	focus := m.state.FocusContext
	if focus.Screen == "" {
		focus.Screen = m.router.Current()
	}
	question := explainQuestionForKind(focus.Kind, focus.Subject)
	return bus.ExplainContext{
		Focus:    focus,
		Question: question,
	}
}

func (m model) explainSubject() string {
	if subject := strings.TrimSpace(m.state.FocusContext.Subject); subject != "" {
		return subject
	}
	return m.router.Current()
}

func (m *model) openExplainForFocus(kind bus.FocusContextKind, subject, detail string) {
	m.setFocusContext(kind, subject, detail)
	m.navigateTo(ScreenExplain, `asa explain --subject "`+explainQuestionForKind(kind, subject)+`"`)
}

func explainQuestionForKind(kind bus.FocusContextKind, subject string) string {
	switch kind {
	case bus.FocusKindGraphNode:
		return "Why is this node blocked?"
	case bus.FocusKindFlowStep:
		return "Why is this flow high risk?"
	case bus.FocusKindTrustDimension:
		if strings.EqualFold(subject, "Security") {
			return "Why is security confidence low?"
		}
		return "Why is security confidence low?"
	case bus.FocusKindAgent:
		return "Why was this agent selected?"
	case bus.FocusKindReplayEvent:
		return "Why did Asagiri insert investigation?"
	default:
		return "Why was review required?"
	}
}
