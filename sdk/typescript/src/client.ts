import type {
  AsagiriClientOptions,
  DaemonStatus,
  MemoryEntry,
  RuntimeEvent,
  Session,
} from "./types.js";

export class AsagiriClient {
  private readonly baseUrl: string;
  private readonly token?: string;
  private readonly fetchFn: typeof fetch;

  constructor(opts: AsagiriClientOptions = {}) {
    this.baseUrl = (opts.baseUrl ?? "http://127.0.0.1:8765").replace(/\/$/, "");
    this.token = opts.token;
    this.fetchFn = opts.fetch ?? fetch;
  }

  async status(): Promise<DaemonStatus> {
    return this.request<DaemonStatus>("GET", "/v1/status");
  }

  async startSession(
    name: string,
    productId = "",
    flowId = ""
  ): Promise<Session> {
    return this.request<Session>("POST", "/v1/sessions", {
      name,
      product_id: productId,
      flow_id: flowId,
    });
  }

  async emitEvent(
    type: string,
    sessionId = "",
    flowId = "",
    payload?: Record<string, unknown>
  ): Promise<RuntimeEvent> {
    return this.request<RuntimeEvent>("POST", "/v1/events", {
      type,
      session_id: sessionId,
      flow_id: flowId,
      payload,
    });
  }

  async runFlow(sessionId: string, flowId: string): Promise<void> {
    await this.emitEvent("flow.started", sessionId, flowId);
    await this.emitEvent("flow.completed", sessionId, flowId);
  }

  async listMemory(scope?: string, limit = 50): Promise<MemoryEntry[]> {
    const q = new URLSearchParams();
    if (scope) q.set("scope", scope);
    q.set("limit", String(limit));
    const res = await this.request<{ memory: MemoryEntry[] }>(
      "GET",
      `/v1/memory?${q}`
    );
    return res.memory ?? [];
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const headers: Record<string, string> = {};
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }
    const res = await this.fetchFn(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
    const text = await res.text();
    if (!res.ok) {
      throw new Error(`runtime api ${method} ${path}: ${res.status} ${text}`);
    }
    return text ? (JSON.parse(text) as T) : ({} as T);
  }
}
