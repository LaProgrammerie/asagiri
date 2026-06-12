CREATE TABLE IF NOT EXISTS run_metrics (
  run_id TEXT PRIMARY KEY,
  feature TEXT,
  task_id TEXT,
  started_at TEXT,
  finished_at TEXT,
  estimated_input_tokens INTEGER,
  estimated_output_tokens INTEGER,
  actual_input_tokens INTEGER,
  actual_output_tokens INTEGER,
  estimated_cost_cents INTEGER,
  actual_cost_cents INTEGER,
  estimated_duration_ms INTEGER,
  actual_duration_ms INTEGER,
  status TEXT
);

CREATE TABLE IF NOT EXISTS step_metrics (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  step_name TEXT,
  agent TEXT,
  model TEXT,
  local BOOLEAN,
  estimated_input_tokens INTEGER,
  estimated_output_tokens INTEGER,
  actual_input_tokens INTEGER,
  actual_output_tokens INTEGER,
  estimated_cost_cents INTEGER,
  actual_cost_cents INTEGER,
  estimated_duration_ms INTEGER,
  actual_duration_ms INTEGER,
  status TEXT
);

CREATE INDEX IF NOT EXISTS idx_step_metrics_run ON step_metrics (run_id);
