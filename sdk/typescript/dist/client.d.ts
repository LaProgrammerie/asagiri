import type { AsagiriClientOptions, DaemonStatus, MemoryEntry, RuntimeEvent, Session } from "./types.js";
export declare class AsagiriClient {
    private readonly baseUrl;
    private readonly token?;
    private readonly fetchFn;
    constructor(opts?: AsagiriClientOptions);
    status(): Promise<DaemonStatus>;
    startSession(name: string, productId?: string, flowId?: string): Promise<Session>;
    emitEvent(type: string, sessionId?: string, flowId?: string, payload?: Record<string, unknown>): Promise<RuntimeEvent>;
    runFlow(sessionId: string, flowId: string): Promise<void>;
    listMemory(scope?: string, limit?: number): Promise<MemoryEntry[]>;
    private request;
}
