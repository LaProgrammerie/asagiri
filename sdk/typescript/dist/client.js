export class AsagiriClient {
    baseUrl;
    token;
    fetchFn;
    constructor(opts = {}) {
        this.baseUrl = (opts.baseUrl ?? "http://127.0.0.1:8765").replace(/\/$/, "");
        this.token = opts.token;
        this.fetchFn = opts.fetch ?? fetch;
    }
    async status() {
        return this.request("GET", "/v1/status");
    }
    async startSession(name, productId = "", flowId = "") {
        return this.request("POST", "/v1/sessions", {
            name,
            product_id: productId,
            flow_id: flowId,
        });
    }
    async emitEvent(type, sessionId = "", flowId = "", payload) {
        return this.request("POST", "/v1/events", {
            type,
            session_id: sessionId,
            flow_id: flowId,
            payload,
        });
    }
    async runFlow(sessionId, flowId) {
        await this.emitEvent("flow.started", sessionId, flowId);
        await this.emitEvent("flow.completed", sessionId, flowId);
    }
    async listMemory(scope, limit = 50) {
        const q = new URLSearchParams();
        if (scope)
            q.set("scope", scope);
        q.set("limit", String(limit));
        const res = await this.request("GET", `/v1/memory?${q}`);
        return res.memory ?? [];
    }
    async request(method, path, body) {
        const headers = {};
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
        return text ? JSON.parse(text) : {};
    }
}
