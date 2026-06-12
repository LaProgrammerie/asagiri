package doctor

const ArchitectureReportVersion = "doctor-architecture-v1"

// ArchitectureReport cross-reads tasks, execution graphs, knowledge graph, trust and agent ledger.
type ArchitectureReport struct {
	ReportVersion   string                `json:"report_version"`
	Repository      ArchitectureRepoInfo  `json:"repository"`
	Summary         ArchitectureSummary   `json:"summary"`
	Sources         ArchitectureSources   `json:"sources"`
	Findings        []ArchitectureFinding `json:"findings"`
	Recommendations []Action              `json:"recommendations"`
}

type ArchitectureRepoInfo struct {
	GitRoot string `json:"git_root,omitempty"`
}

type ArchitectureSummary struct {
	Tasks                   int `json:"tasks"`
	ExecutionGraphs         int `json:"execution_graphs"`
	ExecutionGraphNodes     int `json:"execution_graph_nodes"`
	KnowledgeNodes          int `json:"knowledge_nodes"`
	TrustReports            int `json:"trust_reports"`
	AgentLedgerEntries      int `json:"agent_ledger_entries"`
	TasksWithoutGraphNode   int `json:"tasks_without_graph_node"`
	GraphNodesNeverExecuted int `json:"graph_nodes_never_executed"`
	TasksWithoutKnowledge   int `json:"tasks_without_knowledge_context"`
	AgentRunsWithoutTask    int `json:"agent_runs_without_task"`
	TrustGapsCriticalFlows  int `json:"trust_gaps_critical_flows"`
}

type ArchitectureSources struct {
	SQLitePresent       bool `json:"sqlite_present"`
	GraphsPresent       bool `json:"graphs_present"`
	KnowledgeStore      bool `json:"knowledge_store_present"`
	KnowledgeJSON       bool `json:"knowledge_json_present"`
	TrustReportsPresent bool `json:"trust_reports_present"`
	AgentLedgerPresent  bool `json:"agent_ledger_present"`
}

// ArchitectureFinding is one cross-artefact gap (read-only diagnostic).
type ArchitectureFinding struct {
	Kind     string `json:"kind"`
	Severity string `json:"severity"` // warn | info
	TaskID   string `json:"task_id,omitempty"`
	GraphID  string `json:"graph_id,omitempty"`
	NodeID   string `json:"node_id,omitempty"`
	Flow     string `json:"flow,omitempty"`
	Product  string `json:"product,omitempty"`
	Feature  string `json:"feature,omitempty"`
	RunID    string `json:"run_id,omitempty"`
	AgentID  string `json:"agent_id,omitempty"`
	Message  string `json:"message"`
}
