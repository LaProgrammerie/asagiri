package coordination

import "sync"

// AssignmentHistory records agent outcomes for scoring enrichment (spec-my-D §6).
type AssignmentHistory interface {
	SuccessRate(agentRef string, role AgentRole) float64
	RecordSuccess(agentRef string, role AgentRole)
	RecordFailure(agentRef string, role AgentRole)
}

// MemoryAssignmentHistory keeps in-process success rates per agent+role key.
type MemoryAssignmentHistory struct {
	mu    sync.Mutex
	stats map[string]assignmentStats
}

type assignmentStats struct {
	success int
	total   int
}

func historyKey(agentRef string, role AgentRole) string {
	return string(role) + ":" + agentRef
}

// SuccessRate returns successes / attempts in [0,1], or 0 when unknown.
func (h *MemoryAssignmentHistory) SuccessRate(agentRef string, role AgentRole) float64 {
	if h == nil {
		return 0
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.stats == nil {
		return 0
	}
	st, ok := h.stats[historyKey(agentRef, role)]
	if !ok || st.total == 0 {
		return 0
	}
	return float64(st.success) / float64(st.total)
}

// RecordSuccess increments success and total for agent+role.
func (h *MemoryAssignmentHistory) RecordSuccess(agentRef string, role AgentRole) {
	h.record(agentRef, role, true)
}

// RecordFailure increments total only.
func (h *MemoryAssignmentHistory) RecordFailure(agentRef string, role AgentRole) {
	h.record(agentRef, role, false)
}

func (h *MemoryAssignmentHistory) record(agentRef string, role AgentRole, success bool) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.stats == nil {
		h.stats = make(map[string]assignmentStats)
	}
	key := historyKey(agentRef, role)
	st := h.stats[key]
	st.total++
	if success {
		st.success++
	}
	h.stats[key] = st
}
