export interface DaemonStatus {
  running: boolean;
  pid?: number;
  sessions: number;
  flows_active: number;
  queued_events: number;
  memory_size: number;
  db_path: string;
  db_size_bytes: number;
}

export interface Session {
  id: string;
  name: string;
  product_id?: string;
  flow_id?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface RuntimeEvent {
  id: string;
  type: string;
  session_id?: string;
  flow_id?: string;
  payload?: Record<string, unknown>;
  created_at: string;
}

export interface MemoryEntry {
  id: string;
  scope: string;
  type: string;
  summary: string;
  relevance: number;
}

export interface AsagiriClientOptions {
  baseUrl?: string;
  token?: string;
  fetch?: typeof fetch;
}
