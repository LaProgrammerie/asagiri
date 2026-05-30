package app

const (
	ScreenMission   = "mission"
	ScreenDashboard = "dashboard"
	ScreenAgents    = "agents"
	ScreenGraph     = "graph"
	ScreenFlow      = "flow"
	ScreenLogs      = "logs"
	ScreenExplain   = "explain"
	ScreenReplay    = "replay"
	ScreenPrototype = "prototype"
	ScreenKnowledge = "knowledge"
	ScreenTrust     = "trust"
	ScreenSettings  = "settings"
	ScreenOnboarding = "onboarding"
)

type router struct {
	current string
}

func newRouter(initial string) router {
	r := router{current: ScreenMission}
	r.Set(initial)
	return r
}

func (r *router) Set(screen string) {
	switch screen {
	case ScreenMission, ScreenDashboard, ScreenAgents, ScreenGraph, ScreenFlow, ScreenLogs, ScreenExplain, ScreenReplay, ScreenPrototype, ScreenKnowledge, ScreenTrust, ScreenSettings, ScreenOnboarding:
		r.current = screen
	default:
		r.current = ScreenMission
	}
}

func (r router) Current() string {
	return r.current
}
