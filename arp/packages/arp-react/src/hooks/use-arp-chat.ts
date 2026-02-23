import { useState, useEffect, useRef, useCallback } from "react";
import {
  ArpClient,
  type ArpClientCallbacks,
  type ArpCapabilities,
  type ToolRenderEvent,
  type ChatSession,
  type ChatSessionDetail,
  streamSSE,
  createSessionAPI,
} from "@haira/arp";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  file?: string;
  uiEvents?: ToolRenderEvent[];
  restored?: boolean;
}

export interface ToolCardState {
  name: string;
  displayName: string;
  status: "running" | "done" | "failed";
  startTime: number;
  elapsed?: string;
}

export interface UseArpChatOptions {
  /** WebSocket URL (e.g., "ws://localhost:8080/_arp/v1"). */
  url: string;
  /** API base URL for session persistence (e.g., "http://localhost:8080"). */
  apiBaseUrl?: string;
  /** Initial session ID. If not provided, one is generated. */
  sessionId?: string;
  /** Param name for chat messages (default: "message"). */
  chatParam?: string;
  /** SSE endpoint path for fallback (e.g., "/chat"). */
  sseFallbackPath?: string;
  /** Whether the workflow accepts file uploads. */
  hasFileUpload?: boolean;
  /** Param name for file uploads. */
  fileParam?: string;
}

export interface UseArpChatReturn {
  /** Chat messages in display order. */
  messages: ChatMessage[];
  /** Whether the assistant is currently generating a response. */
  isStreaming: boolean;
  /** Whether the WebSocket is connected. */
  isConnected: boolean;
  /** Active tool executions. */
  toolCards: ToolCardState[];
  /** Count of currently running tools. */
  runningToolCount: number;
  /** Current session ID. */
  sessionId: string;
  /** Server capabilities (null until connected). */
  capabilities: ArpCapabilities | null;

  /** Send a message (optionally with a file attachment). */
  sendMessage: (text: string, file?: File) => void;
  /** Switch to an existing session. */
  switchSession: (sessionId: string) => Promise<void>;
  /** Start a new empty chat session. */
  startNewChat: () => void;

  /** Available sessions for the sidebar. */
  sessions: ChatSession[];
  /** Refresh the sessions list. */
  loadSessions: () => Promise<void>;
  /** Delete a session. */
  deleteSession: (id: string) => Promise<void>;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

let messageCounter = 0;
function nextId(): string {
  return `msg_${++messageCounter}_${Date.now()}`;
}

function formatToolName(name: string): string {
  return name
    .replace(/^render_/, "")
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

/**
 * Chat-specific ARP hook. Manages messages, streaming, tool activity,
 * and session persistence. This is the primary hook for building chat UIs.
 *
 * For headless usage (fully custom rendering):
 * ```tsx
 * const { messages, sendMessage, isStreaming } = useArpChat({
 *   url: "ws://localhost:8080/_arp/v1",
 * });
 * ```
 */
export function useArpChat(options: UseArpChatOptions): UseArpChatReturn {
  const {
    url,
    apiBaseUrl,
    sessionId: initialSessionId,
    chatParam = "message",
    sseFallbackPath,
    hasFileUpload = false,
    fileParam = "file",
  } = options;

  // --- State ---
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [toolCards, setToolCards] = useState<ToolCardState[]>([]);
  const [sessionId, setSessionId] = useState(
    () => initialSessionId || crypto.randomUUID(),
  );
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [capabilities, setCapabilities] = useState<ArpCapabilities | null>(
    null,
  );

  // --- Refs ---
  const clientRef = useRef<ArpClient | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const fullTextRef = useRef("");
  const streamingIdRef = useRef<string | null>(null);

  // --- Session API ---
  const sessionAPI = apiBaseUrl ? createSessionAPI(apiBaseUrl) : null;

  // --- Derived ---
  const runningToolCount = toolCards.filter(
    (t) => t.status === "running",
  ).length;

  // Helper: compute API base from WS URL if not provided
  const effectiveApiBase = apiBaseUrl ?? url.replace(/^ws/, "http").replace(/\/_arp\/v1$/, "");

  // --- Streaming callbacks (shared between ARP and SSE) ---
  const handleDelta = useCallback((delta: string) => {
    fullTextRef.current += delta;
    const text = fullTextRef.current;
    setMessages((prev) => {
      const id = streamingIdRef.current;
      if (!id) return prev;
      return prev.map((m) => (m.id === id ? { ...m, content: text } : m));
    });
  }, []);

  const handleToolStart = useCallback((tool: string, _args: string) => {
    setToolCards((prev) => [
      ...prev,
      {
        name: tool,
        displayName: formatToolName(tool),
        status: "running",
        startTime: Date.now(),
      },
    ]);
  }, []);

  const handleToolEnd = useCallback((tool: string, ok: boolean) => {
    setToolCards((prev) =>
      prev.map((t) =>
        t.name === tool && t.status === "running"
          ? {
              ...t,
              status: ok ? "done" : "failed",
              elapsed: `${((Date.now() - t.startTime) / 1000).toFixed(1)}s`,
            }
          : t,
      ),
    );
  }, []);

  const handleRender = useCallback((event: ToolRenderEvent) => {
    const id = streamingIdRef.current;
    if (!id) return;
    setMessages((prev) =>
      prev.map((m) =>
        m.id === id
          ? { ...m, uiEvents: [...(m.uiEvents ?? []), event] }
          : m,
      ),
    );
  }, []);

  const handleError = useCallback((error: string) => {
    const errId = nextId();
    setMessages((prev) => [
      ...prev,
      { id: errId, role: "assistant", content: `Error: ${error}` },
    ]);
  }, []);

  const handleDone = useCallback(() => {
    setIsStreaming(false);
    setToolCards([]);
    fullTextRef.current = "";
    streamingIdRef.current = null;
  }, []);

  // --- Connect ARP WebSocket ---
  useEffect(() => {
    const callbacks: ArpClientCallbacks = {
      onConnect: (caps) => {
        setIsConnected(true);
        setCapabilities(caps);
      },
      onDisconnect: () => setIsConnected(false),
      onDelta: handleDelta,
      onToolStart: handleToolStart,
      onToolEnd: handleToolEnd,
      onRender: handleRender,
      onError: handleError,
      onDone: handleDone,
    };

    const client = new ArpClient({ url, sessionId }, callbacks);
    clientRef.current = client;
    client.connect();

    return () => {
      client.disconnect();
      clientRef.current = null;
    };
  }, [url, sessionId, handleDelta, handleToolStart, handleToolEnd, handleRender, handleError, handleDone]);

  // --- Load session history on mount ---
  useEffect(() => {
    if (!sessionAPI) return;
    sessionAPI.getSession(sessionId).then((detail) => {
      if (!detail?.messages?.length) return;
      const restored: ChatMessage[] = detail.messages.map((m) => ({
        id: nextId(),
        role: m.role,
        content: m.content,
        uiEvents: m.ui_events,
        restored: true,
      }));
      setMessages(restored);
    });
  }, [sessionId]); // eslint-disable-line react-hooks/exhaustive-deps

  // --- Send message ---
  const sendMessage = useCallback(
    (text: string, file?: File) => {
      if (!text.trim() && !file) return;
      if (isStreaming) return;

      // Add user message
      const userMsg: ChatMessage = {
        id: nextId(),
        role: "user",
        content: text,
        file: file?.name,
      };

      // Add empty assistant message for streaming
      const assistantId = nextId();
      streamingIdRef.current = assistantId;
      fullTextRef.current = "";

      const assistantMsg: ChatMessage = {
        id: assistantId,
        role: "assistant",
        content: "",
      };

      setMessages((prev) => [...prev, userMsg, assistantMsg]);
      setIsStreaming(true);
      setToolCards([]);

      // Try ARP WebSocket (no file support over WS)
      const client = clientRef.current;
      if (client?.connected && !file) {
        client.sendText(text);
        return;
      }

      // SSE fallback
      if (sseFallbackPath) {
        const sseUrl = `${effectiveApiBase}${sseFallbackPath}`;
        const abort = new AbortController();
        abortRef.current = abort;

        let body: unknown;
        let formData: FormData | undefined;

        if (file && hasFileUpload) {
          formData = new FormData();
          formData.append(chatParam, text);
          formData.append(fileParam, file);
          formData.append("session_id", sessionId);
        } else {
          body = { [chatParam]: text, session_id: sessionId };
        }

        streamSSE(sseUrl, body, {
          onDelta: handleDelta,
          onToolStart: (e) => handleToolStart(e.tool, e.args ?? ""),
          onToolEnd: (e) => handleToolEnd(e.tool, e.ok ?? false),
          onToolRender: handleRender,
          onError: handleError,
          onDone: handleDone,
        }, { formData, signal: abort.signal });
      }
    },
    [
      isStreaming, sessionId, chatParam, fileParam, hasFileUpload,
      sseFallbackPath, effectiveApiBase,
      handleDelta, handleToolStart, handleToolEnd, handleRender,
      handleError, handleDone,
    ],
  );

  // --- Session management ---
  const loadSessions = useCallback(async () => {
    if (!sessionAPI) return;
    const list = await sessionAPI.listSessions("");
    setSessions(list);
  }, [sessionAPI]); // eslint-disable-line react-hooks/exhaustive-deps

  const switchSession = useCallback(
    async (newSessionId: string) => {
      // Abort any in-progress stream
      abortRef.current?.abort();
      setIsStreaming(false);
      setToolCards([]);
      fullTextRef.current = "";
      streamingIdRef.current = null;
      setMessages([]);
      setSessionId(newSessionId);
    },
    [],
  );

  const startNewChat = useCallback(() => {
    const newId = crypto.randomUUID();
    switchSession(newId);
  }, [switchSession]);

  const deleteSession = useCallback(
    async (id: string) => {
      if (!sessionAPI) return;
      await sessionAPI.deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
      if (id === sessionId) {
        startNewChat();
      }
    },
    [sessionAPI, sessionId, startNewChat], // eslint-disable-line react-hooks/exhaustive-deps
  );

  return {
    messages,
    isStreaming,
    isConnected,
    toolCards,
    runningToolCount,
    sessionId,
    capabilities,
    sendMessage,
    switchSession,
    startNewChat,
    sessions,
    loadSessions,
    deleteSession,
  };
}
