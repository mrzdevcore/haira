/**
 * ARP Session API — helpers for chat session management.
 *
 * Provides a typed client for the standard ARP session/chat endpoints:
 *   GET  /_api/chats?workflow=<path>  — list sessions
 *   GET  /_api/chats/<id>             — get session detail
 *   DELETE /_api/chats/<id>           — delete session
 *   GET  /_api/workflows              — list available workflows
 */

import type { ChatSession, ChatSessionDetail } from "./types.js";

// ---------------------------------------------------------------------------
// Workflow discovery
// ---------------------------------------------------------------------------

export interface WorkflowInfo {
  name: string;
  path: string;
  method: string;
  is_stream: boolean;
  chat_param?: string;
  title?: string;
  description?: string;
  params?: Array<{ Name: string; Type: string }>;
  suggestions?: string[];
}

// ---------------------------------------------------------------------------
// Session API Client
// ---------------------------------------------------------------------------

export interface SessionAPI {
  /** List sessions for a given workflow path. */
  listSessions(workflowPath: string, owner?: string): Promise<ChatSession[]>;
  /** Get a session with full message history. */
  getSession(sessionId: string): Promise<ChatSessionDetail | null>;
  /** Delete a session. Returns true if successful. */
  deleteSession(sessionId: string): Promise<boolean>;
  /** List available workflows. */
  listWorkflows(): Promise<WorkflowInfo[]>;
}

/**
 * Create a session API client for the given base URL.
 *
 * @param baseUrl - The server base URL (e.g., "http://localhost:8080").
 */
export function createSessionAPI(baseUrl: string): SessionAPI {
  const base = baseUrl.replace(/\/$/, "");

  return {
    async listSessions(
      workflowPath: string,
      owner?: string,
    ): Promise<ChatSession[]> {
      const params = new URLSearchParams({ workflow: workflowPath });
      if (owner) params.set("owner", owner);
      const resp = await fetch(`${base}/_api/chats?${params}`);
      if (!resp.ok) return [];
      return resp.json();
    },

    async getSession(sessionId: string): Promise<ChatSessionDetail | null> {
      const resp = await fetch(`${base}/_api/chats/${sessionId}`);
      if (!resp.ok) return null;
      return resp.json();
    },

    async deleteSession(sessionId: string): Promise<boolean> {
      const resp = await fetch(`${base}/_api/chats/${sessionId}`, {
        method: "DELETE",
      });
      return resp.ok;
    },

    async listWorkflows(): Promise<WorkflowInfo[]> {
      const resp = await fetch(`${base}/_api/workflows`);
      if (!resp.ok) return [];
      return resp.json();
    },
  };
}
