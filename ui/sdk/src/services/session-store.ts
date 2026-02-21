// Session store — handles chat session CRUD operations against the backend API.
// Centralizes all session-related fetch calls that were scattered across components.

import type { ChatSessionSummary, ChatSessionDetail } from "../core/types";

/** List all chat sessions for a given workflow path. */
export async function listSessions(workflowPath: string): Promise<ChatSessionSummary[]> {
  try {
    const resp = await fetch(`/_api/chats?workflow=${encodeURIComponent(workflowPath)}`);
    if (!resp.ok) return [];
    const sessions = await resp.json();
    return sessions || [];
  } catch {
    return [];
  }
}

/** Get the full detail (messages + UI events) for a session. */
export async function getSession(sessionId: string): Promise<ChatSessionDetail | null> {
  try {
    const resp = await fetch(`/_api/chats/${sessionId}`);
    if (!resp.ok) return null;
    return await resp.json();
  } catch {
    return null;
  }
}

/** Delete a session by ID. */
export async function deleteSession(sessionId: string): Promise<boolean> {
  try {
    const resp = await fetch(`/_api/chats/${sessionId}`, { method: "DELETE" });
    return resp.ok;
  } catch {
    return false;
  }
}
