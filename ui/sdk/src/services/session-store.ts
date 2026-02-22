import type { ChatSessionSummary, ChatSessionDetail } from "../core/types";

export async function listSessions(
  workflowPath: string
): Promise<ChatSessionSummary[]> {
  try {
    const resp = await fetch(
      `/_api/chats?workflow=${encodeURIComponent(workflowPath)}`
    );
    if (!resp.ok) return [];
    const sessions = await resp.json();
    return sessions || [];
  } catch {
    return [];
  }
}

export async function getSession(
  sessionId: string
): Promise<ChatSessionDetail | null> {
  try {
    const resp = await fetch(`/_api/chats/${sessionId}`);
    if (!resp.ok) return null;
    return await resp.json();
  } catch {
    return null;
  }
}

export async function deleteSession(sessionId: string): Promise<boolean> {
  try {
    const resp = await fetch(`/_api/chats/${sessionId}`, { method: "DELETE" });
    return resp.ok;
  } catch {
    return false;
  }
}
