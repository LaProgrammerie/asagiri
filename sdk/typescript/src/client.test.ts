import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { AsagiriClient } from "./client.js";
import type { DaemonStatus, Session } from "./types.js";

describe("AsagiriClient", () => {
  it("calls status endpoint", async () => {
    const mockFetch: typeof fetch = async (input, init) => {
      assert.equal(init?.method ?? "GET", "GET");
      assert.match(String(input), /\/v1\/status$/);
      const body: DaemonStatus = {
        running: true,
        sessions: 1,
        flows_active: 0,
        queued_events: 0,
        memory_size: 0,
        db_path: "/tmp/runtime.db",
        db_size_bytes: 100,
      };
      return new Response(JSON.stringify(body), { status: 200 });
    };
    const client = new AsagiriClient({
      baseUrl: "http://127.0.0.1:9999",
      fetch: mockFetch,
    });
    const st = await client.status();
    assert.equal(st.sessions, 1);
  });

  it("creates session", async () => {
    const mockFetch: typeof fetch = async (input, init) => {
      assert.equal(init?.method, "POST");
      assert.match(String(input), /\/v1\/sessions$/);
      const sess: Session = {
        id: "s1",
        name: "demo",
        status: "active",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      };
      return new Response(JSON.stringify(sess), { status: 201 });
    };
    const client = new AsagiriClient({ fetch: mockFetch });
    const sess = await client.startSession("demo", "p1", "flow1");
    assert.equal(sess.id, "s1");
  });
});
